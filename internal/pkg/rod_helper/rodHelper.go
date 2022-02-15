package rod_helper

import (
	"context"
	_ "embed"
	"errors"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/random_useragent"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/mholt/archiver/v3"
	"os"
	"path/filepath"
	"sync"
	"time"
)

/**
 * @Description: 			新建一个支持代理的 browser 对象，使用完毕后，需要删除 adblockFilePath 文件夹
 * @param httpProxyURL		http://127.0.0.1:10809
 * @return *rod.Browser
 * @return error
 */
func NewBrowser(httpProxyURL string, loadAdblock bool) (*rod.Browser, error) {

	var err error

	once.Do(func() {
		adblockSavePath, err = releaseAdblock()
		if err != nil {
			log_helper.GetLogger().Errorln("releaseAdblock", err)
		}
	})
	var browser *rod.Browser
	err = rod.Try(func() {
		purl := ""
		if loadAdblock == true {
			purl = launcher.New().
				Delete("disable-extensions").
				Set("load-extension", adblockSavePath).
				Proxy(httpProxyURL).
				Headless(false). // 插件模式需要设置这个
				//XVFB("--server-num=5", "--server-args=-screen 0 1600x900x16").
				//XVFB("-ac :99", "-screen 0 1280x1024x16").
				MustLaunch()
		} else {
			purl = launcher.New().
				Proxy(httpProxyURL).
				MustLaunch()
		}

		browser = rod.New().ControlURL(purl).MustConnect()
	})
	if err != nil {
		return nil, err
	}

	return browser, nil
}

/**
 * @Description: 			访问目标 Url，返回 page，只是这个 page 有效，如果再次出发其他的事件无效
 * @param desURL			目标 Url
 * @param httpProxyURL		http://127.0.0.1:10809
 * @param timeOut			超时时间
 * @param maxRetryTimes		当是非超时 err 的时候，最多可以重试几次
 * @return *rod.Page
 * @return error
 */
func NewBrowserFromDocker(httpProxyURL, remoteDockerURL string) (*rod.Browser, error) {
	var browser *rod.Browser

	err := rod.Try(func() {
		l := launcher.MustNewManaged(remoteDockerURL)
		u := l.Proxy(httpProxyURL).MustLaunch()
		l.Headless(false).XVFB()
		browser = rod.New().Client(l.Client()).ControlURL(u).MustConnect()
	})
	if err != nil {
		return nil, err
	}

	return browser, nil
}

func NewPageNavigate(browser *rod.Browser, desURL string, timeOut time.Duration, maxRetryTimes int) (*rod.Page, error) {

	page, err := newPage(browser)
	if err != nil {
		return nil, err
	}
	err = page.SetUserAgent(&proto.NetworkSetUserAgentOverride{
		UserAgent: random_useragent.RandomUserAgent(true),
	})
	if err != nil {
		return nil, err
	}
	page = page.Timeout(timeOut)
	nowRetryTimes := 0
	for nowRetryTimes <= maxRetryTimes {
		err = rod.Try(func() {
			page.MustNavigate(desURL).MustWaitLoad()
			nowRetryTimes++
		})
		if errors.Is(err, context.DeadlineExceeded) {
			// 超时
			return nil, err
		} else if err == nil {
			// 没有问题
			return page, nil
		}
	}
	return nil, err
}

// ReloadBrowser 提前把浏览器下载好
func ReloadBrowser() {
	newBrowser, err := NewBrowser("", true)
	if err != nil {
		return
	}
	defer func() {
		_ = newBrowser.Close()
	}()
	page, err := NewPageNavigate(newBrowser, "https://www.baidu.com", 30*time.Second, 5)
	if err != nil {
		return
	}
	defer func() {
		_ = page.Close()
	}()
}

// Clear 清理缓存
func Clear() {
	_ = rod.Try(func() {
		l := launcher.New().
			Headless(false).
			Devtools(true)

		defer l.Cleanup() // remove launcher.FlagUserDataDir

		url := l.MustLaunch()
		// Trace shows verbose debug information for each action executed
		// Slowmotion is a debug related function that waits 2 seconds between
		// each action, making it easier to inspect what your code is doing.
		browser := rod.New().
			ControlURL(url).
			Trace(true).
			SlowMotion(2 * time.Second).
			MustConnect()
		defer browser.MustClose()
	})
}

func newPage(browser *rod.Browser) (*rod.Page, error) {
	page, err := browser.Page(proto.TargetCreateTarget{URL: ""})
	if err != nil {
		return nil, err
	}
	return page, err
}

// releaseAdblock 从程序中释放 adblock 插件出来到本地路径
func releaseAdblock() (string, error) {

	adblockFolderPath := filepath.Join(os.TempDir(), "chinesesubfinder")
	err := os.MkdirAll(filepath.Join(adblockFolderPath), os.ModePerm)
	if err != nil {
		return "", err
	}
	desPath := filepath.Join(adblockFolderPath, "RunAdblock")
	// 清理之前缓存的信息
	_ = my_util.ClearFolder(desPath)
	// 具体把 adblock zip 解压下载到哪里
	outZipFileFPath := filepath.Join(adblockFolderPath, "adblock.zip")
	adblockZipFile, err := os.Create(outZipFileFPath)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = adblockZipFile.Close()
		_ = os.Remove(outZipFileFPath)
	}()
	_, err = adblockZipFile.Write(adblockFolder)
	if err != nil {
		return "", err
	}
	_ = adblockZipFile.Close()

	r := archiver.NewZip()
	err = r.Unarchive(outZipFileFPath, desPath)
	if err != nil {
		return "", err
	}
	return filepath.Join(desPath, adblockInsideName), err
}

const adblockInsideName = "adblock"

var once sync.Once

// 这个文件内有一个子文件夹 adblock ，制作的时候务必注意
//go:embed assets/adblock_4_42_0_0.zip
var adblockFolder []byte

var adblockSavePath string
