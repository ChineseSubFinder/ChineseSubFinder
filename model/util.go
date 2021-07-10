package model

import (
	"fmt"
	browser "github.com/EDDYCJY/fake-useragent"
	"github.com/allanpk716/ChineseSubFinder/common"
	"github.com/go-resty/resty/v2"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// NewHttpClient 新建一个 resty 的对象
func NewHttpClient(_reqParam ...common.ReqParam) *resty.Client {
	//const defUserAgent = "Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10_6_8; en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50"
	//const defUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36 Edg/91.0.864.41"
	// 随机的 Browser
	defUserAgent := browser.Random()

	var reqParam common.ReqParam
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
		"User-Agent": UserAgent,
	})
	if len(Referer) > 0 {
		httpClient.SetHeader("Referer", Referer)
	}

	return httpClient
}

// DownFile 从指定的 url 下载文件
func DownFile(urlStr string, _reqParam ...common.ReqParam) ([]byte, string, error)  {
	var reqParam common.ReqParam
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
	tmpFolderFullPath :=path.Join(rootPath, folderName)
	err = os.MkdirAll(tmpFolderFullPath, os.ModePerm)
	if err != nil {
		return "", err
	}
	return tmpFolderFullPath, nil
}

// ClearTmpFolder 清理指定的缓存文件夹
func ClearTmpFolder(folderName string) error {

	nowTmpFolder, err := GetTmpFolder(folderName)
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

func CopyFile(dstName, srcName string) (written int64, err error) {
	src, err := os.Open(srcName)
	if err != nil {
		return
	}
	defer src.Close()

	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	defer dst.Close()
	return io.Copy(dst, src)
}

func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// VideoNameSearchKeywordMaker 拼接视频搜索的 title 和 年份
func VideoNameSearchKeywordMaker(title string, year string) string {
	iYear, err := strconv.Atoi(year)
	if err != nil {
		// 允许的错误
		GetLogger().Errorln("VideoNameSearchKeywordMaker", "year to int", err)
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

// SearchMatchedSubFile 搜索符合后缀名的视频文件
func SearchMatchedSubFile(dir string) ([]string, error) {
	// 这里有个梗，会出现 __MACOSX 这类文件夹，那么里面会有一样的文件，需要用文件大小排除一下，至少大于 1 kb 吧
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
			oneList, _ := SearchMatchedSubFile(fullPath)
			if oneList != nil {
				fileFullPathList = append(fileFullPathList, oneList...)
			}
		} else {
			// 这里就是文件了
			if curFile.Size() < 1000 {
				continue
			}
			if IsSubExtWanted(filepath.Ext(curFile.Name())) == true {
				fileFullPathList = append(fileFullPathList, fullPath)
			}
		}
	}
	return fileFullPathList, nil
}

// IsWantedVideoExtDef 后缀名是否符合规则
func IsWantedVideoExtDef(fileName string) bool {
	// TODO 强制使用固定的视频后缀名匹配列表，后续有需求再考虑额实现外部可配置的列表

	if len(wantedExtList) < 1 {
		defExtList = append(defExtList, common.VideoExtMp4)
		defExtList = append(defExtList, common.VideoExtMkv)
		defExtList = append(defExtList, common.VideoExtRmvb)
		defExtList = append(defExtList, common.VideoExtIso)

		wantedExtList = append(defExtList, common.VideoExtMp4)
		wantedExtList = append(defExtList, common.VideoExtMkv)
		wantedExtList = append(defExtList, common.VideoExtRmvb)
		wantedExtList = append(defExtList, common.VideoExtIso)
	}
	fileName = strings.ToLower(filepath.Ext(fileName))
	for _, s := range wantedExtList {
		if s == fileName {
			return true
		}
	}
	return false
}

func GetEpisodeKeyName(season, eps int) string {
	return "S" + strconv.Itoa(season) + "E" +strconv.Itoa(eps)
}

// ReloadBrowser 提前把浏览器下载好
func ReloadBrowser() {
	page, err := NewBrowserLoadPage("https://www.baidu.com", "", 300*time.Second, 2)
	if err != nil {
		return
	}
	defer page.Close()
}

var (
	defDebugFolder = ""
	defTmpFolder = ""
	wantedExtList = make([]string, 0)                   // 人工确认的需要监控的视频后缀名
	defExtList    = make([]string, 0)                  // 内置支持的视频后缀名列表
)