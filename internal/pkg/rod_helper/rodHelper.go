package rod_helper

import (
	"context"
	_ "embed"
	"errors"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/global_value"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_folder"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/random_useragent"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/mholt/archiver/v3"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func NewBrowserEx(log *logrus.Logger, loadAdblock bool, _settings *settings.Settings, preLoadUrl ...string) (*rod.Browser, error) {

	if _settings.ExperimentalFunction.RemoteChromeSettings.Enable == false {

		localChromeFPath := ""
		if _settings.ExperimentalFunction.LocalChromeSettings.Enabled == true {
			localChromeFPath = _settings.ExperimentalFunction.LocalChromeSettings.LocalChromeExeFPath
		}
		return NewBrowser(log,
			_settings.AdvancedSettings.ProxySettings.GetLocalHttpProxyUrl(),
			localChromeFPath,
			loadAdblock,
			preLoadUrl...)
	} else {
		return NewBrowserFromDocker(_settings.AdvancedSettings.ProxySettings.GetLocalHttpProxyUrl(),
			_settings.ExperimentalFunction.RemoteChromeSettings.RemoteDockerURL,
			_settings.ExperimentalFunction.RemoteChromeSettings.RemoteAdblockPath,
			_settings.ExperimentalFunction.RemoteChromeSettings.ReMoteUserDataDir,
			loadAdblock, preLoadUrl...)
	}
}

func NewBrowser(log *logrus.Logger, localChromeFPath, httpProxyURL string, loadAdblock bool, preLoadUrl ...string) (*rod.Browser, error) {

	var err error

	once.Do(func() {
		adblockSavePath, err = releaseAdblock(log)
		if err != nil {
			log.Errorln("releaseAdblock", err)
			log.Panicln("releaseAdblock", err)
		}
	})

	// 随机的 rod 子文件夹名称
	nowUserData := filepath.Join(global_value.DefRodTmpRootFolder(), my_util.RandStringBytesMaskImprSrcSB(20))
	var browser *rod.Browser

	if localChromeFPath != "" {
		// 如果有指定的 chrome 路径，则使用指定的 chrome 路径
		if my_util.IsFile(localChromeFPath) == false {
			log.Errorln(errors.New("localChromeFPath is not a file"))
			panic(errors.New("localChromeFPath is not a file"))
		}
		err = rod.Try(func() {
			purl := ""
			if loadAdblock == true {
				purl = launcher.New().Bin(localChromeFPath).
					Delete("disable-extensions").
					Set("load-extension", adblockSavePath).
					Proxy(httpProxyURL).
					Headless(false). // 插件模式需要设置这个
					UserDataDir(nowUserData).
					//XVFB("--server-num=5", "--server-args=-screen 0 1600x900x16").
					//XVFB("-ac :99", "-screen 0 1280x1024x16").
					MustLaunch()
			} else {
				purl = launcher.New().Bin(localChromeFPath).
					Proxy(httpProxyURL).
					UserDataDir(nowUserData).
					MustLaunch()
			}

			browser = rod.New().ControlURL(purl).MustConnect()
		})
	} else {
		// 如果没有指定 chrome 的路径，则使用 rod 自行下载的 chrome
		err = rod.Try(func() {
			purl := ""
			if loadAdblock == true {
				purl = launcher.New().
					Delete("disable-extensions").
					Set("load-extension", adblockSavePath).
					Proxy(httpProxyURL).
					Headless(false). // 插件模式需要设置这个
					UserDataDir(nowUserData).
					//XVFB("--server-num=5", "--server-args=-screen 0 1600x900x16").
					//XVFB("-ac :99", "-screen 0 1280x1024x16").
					MustLaunch()
			} else {
				purl = launcher.New().
					Proxy(httpProxyURL).
					UserDataDir(nowUserData).
					MustLaunch()
			}

			browser = rod.New().ControlURL(purl).MustConnect()
		})
	}

	if err != nil {
		return nil, err
	}

	// 如果加载了插件，那么就需要进行一定的耗时操作，等待其第一次的加载完成
	if loadAdblock == true {
		_, _, err := HttpGetFromBrowser(browser, "https://www.qq.com", 15*time.Second)
		if err != nil {
			if browser != nil {
				browser.Close()
			}
			return nil, err
		}

		//if page != nil {
		//	_ = page.Close()
		//}
	}
	if len(preLoadUrl) > 0 {
		_, _, err := HttpGetFromBrowser(browser, preLoadUrl[0], 15*time.Second)
		if err != nil {
			if browser != nil {
				browser.Close()
			}
			return nil, err
		}

		//if page != nil {
		//	_ = page.Close()
		//}
	}

	return browser, nil
}

func NewBrowserFromDocker(httpProxyURL, remoteDockerURL string, remoteAdblockPath, reMoteUserDataDir string,
	loadAdblock bool, preLoadUrl ...string) (*rod.Browser, error) {
	var browser *rod.Browser

	err := rod.Try(func() {

		purl := ""
		var l *launcher.Launcher
		if loadAdblock == true {
			l = launcher.MustNewManaged(remoteDockerURL)
			purl = l.Delete("disable-extensions").
				Set("load-extension", remoteAdblockPath).
				Proxy(httpProxyURL).
				Headless(false). // 插件模式需要设置这个
				UserDataDir(reMoteUserDataDir).
				MustLaunch()
		} else {
			l = launcher.MustNewManaged(remoteDockerURL)
			purl = l.
				Proxy(httpProxyURL).
				UserDataDir(reMoteUserDataDir).
				MustLaunch()
		}

		browser = rod.New().Client(l.MustClient()).ControlURL(purl).MustConnect()
	})
	if err != nil {
		return nil, err
	}

	// 如果加载了插件，那么就需要进行一定的耗时操作，等待其第一次的加载完成
	if loadAdblock == true {
		_, page, err := HttpGetFromBrowser(browser, "https://www.qq.com", 15*time.Second)
		if err != nil {
			if browser != nil {
				browser.Close()
			}
			return nil, err
		}

		if page != nil {
			_ = page.Close()
		}
	}

	if len(preLoadUrl) > 0 {
		_, page, err := HttpGetFromBrowser(browser, preLoadUrl[0], 15*time.Second)
		if err != nil {
			if browser != nil {
				browser.Close()
			}
			return nil, err
		}

		if page != nil {
			_ = page.Close()
		}
	}

	return browser, nil
}

func NewPageNavigate(browser *rod.Browser, desURL string, timeOut time.Duration, maxRetryTimes int) (*rod.Page, error) {

	addSleepTime := time.Second * 5

	page, err := newPage(browser)
	if err != nil {
		return nil, err
	}
	err = page.SetUserAgent(&proto.NetworkSetUserAgentOverride{
		UserAgent: random_useragent.RandomUserAgent(true),
	})
	if err != nil {
		if page != nil {
			page.Close()
		}
		return nil, err
	}
	page = page.Timeout(timeOut + addSleepTime)
	nowRetryTimes := 0
	for nowRetryTimes <= maxRetryTimes {
		err = rod.Try(func() {
			page.MustNavigate(desURL).MustWaitLoad()
			time.Sleep(addSleepTime)
			nowRetryTimes++
		})
		if errors.Is(err, context.DeadlineExceeded) {
			// 超时
			if page != nil {
				page.Close()
			}
			return nil, err
		} else if err == nil {
			// 没有问题
			return page, nil
		}
	}
	if page != nil {
		page.Close()
	}
	return nil, err
}

func HttpGetFromBrowser(browser *rod.Browser, inputUrl string, tt time.Duration, debugMode ...bool) (string, *rod.Page, error) {

	page, err := NewPageNavigate(browser, inputUrl, tt, 2)
	if err != nil {
		return "", nil, err
	}
	pageString, err := page.HTML()
	if err != nil {
		if page != nil {
			page.Close()
		}
		return "", nil, err
	}
	// 每次搜索间隔
	if len(debugMode) > 0 && debugMode[0] == true {
		time.Sleep(my_util.RandomSecondDuration(2, 5))
	} else {
		time.Sleep(my_util.RandomSecondDuration(2, 5))
	}

	if strings.Contains(strings.ToLower(pageString), "<title>403 forbidden</title>") == true {
		if page != nil {
			page.Close()
		}
		return "", nil, errors.New("403 forbidden")
	}

	return pageString, page, nil
}

// ReloadBrowser 提前把浏览器下载好
func ReloadBrowser(log *logrus.Logger) {
	newBrowser, err := NewBrowserEx(log, true, settings.GetSettings())
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
func Clear(log *logrus.Logger) {
	err := my_folder.ClearRodTmpRootFolder()
	if err != nil {
		log.Errorln("ClearRodTmpRootFolder", err)
		return
	}

	log.Infoln("ClearRodTmpRootFolder Done")
}

func newPage(browser *rod.Browser) (*rod.Page, error) {
	page, err := browser.Page(proto.TargetCreateTarget{URL: ""})
	if err != nil {
		return nil, err
	}
	return page, err
}

// releaseAdblock 从程序中释放 adblock 插件出来到本地路径
func releaseAdblock(log *logrus.Logger) (string, error) {

	defer func() {
		log.Infoln("releaseAdblock end")
	}()

	log.Infoln("releaseAdblock start")

	adblockFolderPath := global_value.AdblockTmpFolder()
	err := os.MkdirAll(filepath.Join(adblockFolderPath), os.ModePerm)
	if err != nil {
		return "", err
	}
	desPath := filepath.Join(adblockFolderPath, "RunAdblock")
	// 清理之前缓存的信息
	_ = my_folder.ClearFolder(desPath)
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
//go:embed assets/adblock_4_43_0_0.zip
var adblockFolder []byte

var adblockSavePath string
