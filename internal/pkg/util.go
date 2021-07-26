package pkg

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/common"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/rod_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	browser "github.com/allanpk716/fake-useragent"
	"github.com/go-resty/resty/v2"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// NewHttpClient 新建一个 resty 的对象
func NewHttpClient(_reqParam ...types.ReqParam) *resty.Client {
	//const defUserAgent = "Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10_6_8; en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50"
	//const defUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36 Edg/91.0.864.41"
	// 随机的 Browser
	defUserAgent := browser.Random()

	var reqParam types.ReqParam
	var HttpProxy, UserAgent, Referer string

	if len(_reqParam) > 0 {
		reqParam = _reqParam[0]
	}
	if len(reqParam.HttpProxy) > 0 {
		HttpProxy = reqParam.HttpProxy
	}
	if len(reqParam.UserAgent) > 0 {
		UserAgent = reqParam.UserAgent
	} else {
		UserAgent = defUserAgent
	}
	if len(reqParam.Referer) > 0 {
		Referer = reqParam.Referer
	}

	httpClient := resty.New()
	httpClient.SetTimeout(common.HTMLTimeOut)
	if HttpProxy != "" {
		httpClient.SetProxy(HttpProxy)
	} else {
		httpClient.RemoveProxy()
	}

	httpClient.SetHeaders(map[string]string{
		"Content-Type": "application/json",
		"User-Agent":   UserAgent,
	})
	if len(Referer) > 0 {
		httpClient.SetHeader("Referer", Referer)
	}

	return httpClient
}

// DownFile 从指定的 url 下载文件
func DownFile(urlStr string, _reqParam ...types.ReqParam) ([]byte, string, error) {
	var reqParam types.ReqParam
	if len(_reqParam) > 0 {
		reqParam = _reqParam[0]
	}
	httpClient := NewHttpClient(reqParam)
	resp, err := httpClient.R().Get(urlStr)
	if err != nil {
		return nil, "", err
	}
	filename := GetFileName(resp.RawResponse)
	return resp.Body(), filename, nil
}

// GetFileName 获取下载文件的文件名
func GetFileName(resp *http.Response) string {
	contentDisposition := resp.Header.Get("Content-Disposition")
	if len(contentDisposition) == 0 {
		return ""
	}
	re := regexp.MustCompile(`filename=["]*([^"]+)["]*`)
	matched := re.FindStringSubmatch(contentDisposition)
	if matched == nil || len(matched) == 0 || len(matched[0]) == 0 {
		//fmt.Println("######")
		return ""
	}
	return matched[1]
}

// AddBaseUrl 判断传入的 url 是否需要拼接 baseUrl
func AddBaseUrl(baseUrl, url string) string {
	if strings.Contains(url, "://") {
		return url
	}
	return fmt.Sprintf("%s%s", baseUrl, url)
}

func GetDebugFolder() (string, error) {
	if defDebugFolder == "" {
		nowProcessRoot, _ := os.Getwd()
		nowProcessRoot = path.Join(nowProcessRoot, common.DebugFolder)
		err := os.MkdirAll(nowProcessRoot, os.ModePerm)
		if err != nil {
			return "", err
		}
		defDebugFolder = nowProcessRoot
		return nowProcessRoot, err
	}
	return defDebugFolder, nil
}

// GetRootTmpFolder 获取缓存的根目录，每一个视频的缓存将在其中额外新建子集文件夹
func GetRootTmpFolder() (string, error) {
	if defTmpFolder == "" {
		nowProcessRoot, _ := os.Getwd()
		nowProcessRoot = path.Join(nowProcessRoot, common.TmpFolder)
		err := os.MkdirAll(nowProcessRoot, os.ModePerm)
		if err != nil {
			return "", err
		}
		defTmpFolder = nowProcessRoot
		return nowProcessRoot, err
	}
	return defTmpFolder, nil
}

// ClearRootTmpFolder 清理缓存的根目录，将里面的子文件夹一并清理
func ClearRootTmpFolder() error {
	nowTmpFolder, err := GetRootTmpFolder()
	if err != nil {
		return err
	}

	pathSep := string(os.PathSeparator)
	files, err := ioutil.ReadDir(nowTmpFolder)
	if err != nil {
		return err
	}
	for _, curFile := range files {
		fullPath := nowTmpFolder + pathSep + curFile.Name()
		if curFile.IsDir() {
			err = os.RemoveAll(fullPath)
			if err != nil {
				return err
			}
		} else {
			// 这里就是文件了
			err = os.Remove(fullPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// GetTmpFolder 获取缓存的文件夹，没有则新建
func GetTmpFolder(folderName string) (string, error) {
	rootPath, err := GetRootTmpFolder()
	if err != nil {
		return "", err
	}
	tmpFolderFullPath := path.Join(rootPath, folderName)
	err = os.MkdirAll(tmpFolderFullPath, os.ModePerm)
	if err != nil {
		return "", err
	}
	return tmpFolderFullPath, nil
}

// ClearFolder 清空文件夹
func ClearFolder(folderName string) error {
	pathSep := string(os.PathSeparator)
	files, err := ioutil.ReadDir(folderName)
	if err != nil {
		return err
	}
	for _, curFile := range files {
		fullPath := folderName + pathSep + curFile.Name()
		if curFile.IsDir() {
			err = os.RemoveAll(fullPath)
			if err != nil {
				return err
			}
		} else {
			// 这里就是文件了
			err = os.Remove(fullPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// ClearTmpFolder 清理指定的缓存文件夹
func ClearTmpFolder(folderName string) error {

	nowTmpFolder, err := GetTmpFolder(folderName)
	if err != nil {
		return err
	}

	return ClearFolder(nowTmpFolder)
}

// IsDir 存在且是文件夹
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// IsFile 存在且是文件
func IsFile(filePath string) bool {
	s, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	return !s.IsDir()
}

// VideoNameSearchKeywordMaker 拼接视频搜索的 title 和 年份
func VideoNameSearchKeywordMaker(title string, year string) string {
	iYear, err := strconv.Atoi(year)
	if err != nil {
		// 允许的错误
		log_helper.GetLogger().Errorln("VideoNameSearchKeywordMaker", "year to int", err)
		iYear = 0
	}
	searchKeyword := title
	if iYear >= 2020 {
		searchKeyword = searchKeyword + " " + year
	}

	return searchKeyword
}

// SearchMatchedVideoFile 搜索符合后缀名的视频文件
func SearchMatchedVideoFile(dir string) ([]string, error) {

	var fileFullPathList = make([]string, 0)
	pathSep := string(os.PathSeparator)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, curFile := range files {
		fullPath := dir + pathSep + curFile.Name()
		if curFile.IsDir() {
			// 内层的错误就无视了
			oneList, _ := SearchMatchedVideoFile(fullPath)
			if oneList != nil {
				fileFullPathList = append(fileFullPathList, oneList...)
			}
		} else {
			// 这里就是文件了
			if IsWantedVideoExtDef(curFile.Name()) == true {
				fileFullPathList = append(fileFullPathList, fullPath)
			}
		}
	}
	return fileFullPathList, nil
}

// IsWantedVideoExtDef 后缀名是否符合规则
func IsWantedVideoExtDef(fileName string) bool {

	if len(wantedExtMap) < 1 {
		defExtMap[common.VideoExtMp4] = common.VideoExtMp4
		defExtMap[common.VideoExtMkv] = common.VideoExtMkv
		defExtMap[common.VideoExtRmvb] = common.VideoExtRmvb
		defExtMap[common.VideoExtIso] = common.VideoExtIso

		wantedExtMap[common.VideoExtMp4] = common.VideoExtMp4
		wantedExtMap[common.VideoExtMkv] = common.VideoExtMkv
		wantedExtMap[common.VideoExtRmvb] = common.VideoExtRmvb
		wantedExtMap[common.VideoExtIso] = common.VideoExtIso

		for _, videoExt := range customVideoExts {
			wantedExtMap[videoExt] = videoExt
		}
	}
	fileExt := strings.ToLower(filepath.Ext(fileName))
	_, bFound := wantedExtMap[fileExt]
	return bFound
}

func GetEpisodeKeyName(season, eps int) string {
	return "S" + strconv.Itoa(season) + "E" + strconv.Itoa(eps)
}

// ReloadBrowser 提前把浏览器下载好
func ReloadBrowser() {
	page, err := rod_helper.NewBrowserLoadPage("https://www.baidu.com", "", 300*time.Second, 2)
	if err != nil {
		return
	}
	defer page.Close()
}

// CopyFile copies a single file from src to dst
func CopyFile(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

// CopyDir copies a whole directory recursively
func CopyDir(src string, dst string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = ioutil.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = CopyDir(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = CopyFile(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

// CopyTestData 单元测试前把测试的数据 copy 一份出来操作，src 目录中默认应该有一个 org 原始数据文件夹，然后需要复制一份 test 文件夹出来
func CopyTestData(srcDir string) (string, error) {
	// 测试数据的文件夹
	orgDir := path.Join(srcDir, "org")
	testDir := path.Join(srcDir, "test")

	if IsDir(testDir) == true {
		err := ClearFolder(testDir)
		if err != nil {
			return "", err
		}
	}

	err := CopyDir(orgDir, testDir)
	if err != nil {
		return "", err
	}
	return testDir, nil
}

// CloseChrome 强行结束没有关闭的 Chrome 进程
func CloseChrome() {

	cmdString := ""
	sysType := runtime.GOOS
	if sysType == "linux" {
		// LINUX系统
		cmdString = "pkill chrome"
	}
	if sysType == "windows" {
		// windows系统
		cmdString = "taskkill /F /im chromedriver.exe"
	}

	if cmdString == "" {
		log_helper.GetLogger().Errorln("CloseChrome OS:", sysType)
		return
	}

	command := exec.Command(cmdString)
	err := command.Run()
	if err != nil {
		log_helper.GetLogger().Errorln("CloseChrome", err)
	}
}
