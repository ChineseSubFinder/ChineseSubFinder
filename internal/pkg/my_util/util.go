package my_util

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/common"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/global_value"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_folder"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/regex_things"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	browser "github.com/allanpk716/fake-useragent"
	"github.com/go-resty/resty/v2"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"os/exec"
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
	httpClient.SetRetryCount(2)
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

	if len(global_value.WantedExtMap) < 1 {
		global_value.DefExtMap[common.VideoExtMp4] = common.VideoExtMp4
		global_value.DefExtMap[common.VideoExtMkv] = common.VideoExtMkv
		global_value.DefExtMap[common.VideoExtRmvb] = common.VideoExtRmvb
		global_value.DefExtMap[common.VideoExtIso] = common.VideoExtIso

		global_value.WantedExtMap[common.VideoExtMp4] = common.VideoExtMp4
		global_value.WantedExtMap[common.VideoExtMkv] = common.VideoExtMkv
		global_value.WantedExtMap[common.VideoExtRmvb] = common.VideoExtRmvb
		global_value.WantedExtMap[common.VideoExtIso] = common.VideoExtIso

		for _, videoExt := range global_value.CustomVideoExts {
			global_value.WantedExtMap[videoExt] = videoExt
		}
	}
	fileExt := strings.ToLower(filepath.Ext(fileName))
	_, bFound := global_value.WantedExtMap[fileExt]
	return bFound
}

func GetEpisodeKeyName(season, eps int) string {
	return "S" + strconv.Itoa(season) + "E" + strconv.Itoa(eps)
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
		srcfp := filepath.Join(src, fd.Name())
		dstfp := filepath.Join(dst, fd.Name())

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
	orgDir := filepath.Join(srcDir, "org")
	testDir := filepath.Join(srcDir, "test")

	if IsDir(testDir) == true {
		err := my_folder.ClearFolder(testDir)
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
	var command *exec.Cmd
	sysType := runtime.GOOS
	if sysType == "linux" {
		// LINUX系统
		cmdString = "pkill chrome"
		command = exec.Command("/bin/sh", "-c", cmdString)
	}
	if sysType == "windows" {
		// windows系统
		cmdString = "taskkill /F /im notepad.exe"
		command = exec.Command("cmd.exe", "/c", cmdString)
	}
	if cmdString == "" || command == nil {
		log_helper.GetLogger().Errorln("CloseChrome OS:", sysType)
		return
	}
	err := command.Run()
	if err != nil {
		log_helper.GetLogger().Errorln("CloseChrome", err)
	}
}

// OSCheck 强制的系统支持检查
func OSCheck() bool {
	sysType := runtime.GOOS
	if sysType == "linux" {
		return true
	}
	if sysType == "windows" {
		return true
	}

	return false
}

// FixWindowPathBackSlash 修复 Windows 反斜杠的梗
func FixWindowPathBackSlash(path string) string {
	return strings.Replace(path, string(filepath.Separator), "/", -1)
}

func WriteStrings2File(desfilePath string, strings []string) error {
	dstFile, err := os.Create(desfilePath)
	if err != nil {
		return err
	}
	defer func() {
		_ = dstFile.Close()
	}()
	allString := ""
	for _, s := range strings {
		allString += s + "\r\n"
	}
	_, err = dstFile.WriteString(allString)
	if err != nil {
		return err
	}
	return nil
}

func Time2SecondNumber(inTime time.Time) float64 {
	outSecond := 0.0
	outSecond += float64(inTime.Hour() * 60 * 60)
	outSecond += float64(inTime.Minute() * 60)
	outSecond += float64(inTime.Second())
	outSecond += float64(inTime.Nanosecond()) / 1000 / 1000 / 1000

	return outSecond
}

func Time2Duration(inTime time.Time) time.Duration {
	return time.Duration(Time2SecondNumber(inTime) * math.Pow10(9))
}

// ReplaceSpecString 替换特殊的字符
func ReplaceSpecString(instring string, rep string) string {
	return regex_things.RegMatchSpString.ReplaceAllString(instring, rep)
}

func Bool2Int(inBool bool) int {
	if inBool == true {
		return 1
	} else {
		return 0
	}
}

// Round 取整
func Round(x float64) int64 {

	if x-float64(int64(x)) > 0 {
		return int64(x) + 1
	} else {
		return int64(x)
	}

	//return int64(math.Floor(x + 0.5))
}

// MakePowerOfTwo 2的整次幂数 buffer length is not a power of two
func MakePowerOfTwo(x int64) int64 {

	power := math.Log2(float64(x))
	tmpRound := Round(power)

	return int64(math.Pow(2, float64(tmpRound)))
}

// MakeCeil10msMultipleFromFloat 将传入的秒，规整到 10ms 的倍数，返回依然是 秒，向上取整
func MakeCeil10msMultipleFromFloat(input float64) float64 {
	const bb = 100
	// 先转到 10 ms 单位，比如传入是 1.912 - > 191.2
	t10ms := input * bb
	// 191.2 - > 192.0
	newT10ms := math.Ceil(t10ms)
	// 转换回来
	return newT10ms / bb
}

// MakeFloor10msMultipleFromFloat 将传入的秒，规整到 10ms 的倍数，返回依然是 秒，向下取整
func MakeFloor10msMultipleFromFloat(input float64) float64 {
	const bb = 100
	// 先转到 10 ms 单位，比如传入是 1.912 - > 191.2
	t10ms := input * bb
	// 191.2 - > 191.0
	newT10ms := math.Floor(t10ms)
	// 转换回来
	return newT10ms / bb
}

// MakeCeil10msMultipleFromTime 向上取整，规整到 10ms 的倍数
func MakeCeil10msMultipleFromTime(input time.Time) time.Time {

	nowTime := MakeCeil10msMultipleFromFloat(Time2SecondNumber(input))
	newTime := time.Time{}.Add(time.Duration(nowTime * math.Pow10(9)))
	return newTime
}

// MakeFloor10msMultipleFromTime 向下取整，规整到 10ms 的倍数
func MakeFloor10msMultipleFromTime(input time.Time) time.Time {

	nowTime := MakeFloor10msMultipleFromFloat(Time2SecondNumber(input))
	newTime := time.Time{}.Add(time.Duration(nowTime * math.Pow10(9)))
	return newTime
}
