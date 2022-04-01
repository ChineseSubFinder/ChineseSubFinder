package my_util

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/common"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/decode"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/regex_things"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	browser "github.com/allanpk716/fake-useragent"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"io"
	"math"
	"net/http"
	"net/url"
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
func NewHttpClient(_proxySettings ...settings.ProxySettings) *resty.Client {
	//const defUserAgent = "Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10_6_8; en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50"
	//const defUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36 Edg/91.0.864.41"

	var proxySettings settings.ProxySettings
	var HttpProxy, UserAgent, Referer string

	if len(_proxySettings) > 0 {
		proxySettings = _proxySettings[0]
	}
	if proxySettings.UseHttpProxy == true && len(proxySettings.HttpProxyAddress) > 0 {
		HttpProxy = proxySettings.HttpProxyAddress
	}
	// 随机的 Browser
	UserAgent = browser.Random()

	httpClient := resty.New()
	httpClient.SetTimeout(common.HTMLTimeOut)
	httpClient.SetRetryCount(2)
	if HttpProxy != "" {
		httpClient.SetProxy(HttpProxy)
	} else {
		httpClient.RemoveProxy()
	}

	if len(proxySettings.Referer) > 0 {
		Referer = proxySettings.Referer
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
func DownFile(l *logrus.Logger, urlStr string, _proxySettings ...settings.ProxySettings) ([]byte, string, error) {
	var proxySettings settings.ProxySettings
	if len(_proxySettings) > 0 {
		proxySettings = _proxySettings[0]
	}
	httpClient := NewHttpClient(proxySettings)
	resp, err := httpClient.R().Get(urlStr)
	if err != nil {
		return nil, "", err
	}
	filename := GetFileName(l, resp.RawResponse)

	if filename == "" {
		l.Warningln("DownFile.GetFileName is string.empty", urlStr)
	}

	return resp.Body(), filename, nil
}

// GetFileName 获取下载文件的文件名
func GetFileName(l *logrus.Logger, resp *http.Response) string {
	contentDisposition := resp.Header.Get("Content-Disposition")
	if len(contentDisposition) == 0 {
		return ""
	}
	re := regexp.MustCompile(`filename=["]*([^"]+)["]*`)
	matched := re.FindStringSubmatch(contentDisposition)
	if matched == nil || len(matched) == 0 || len(matched[0]) == 0 {
		l.Errorln("GetFileName.Content-Disposition", contentDisposition)
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
func VideoNameSearchKeywordMaker(l *logrus.Logger, title string, year string) string {
	iYear, err := strconv.Atoi(year)
	if err != nil {
		// 允许的错误
		l.Errorln("VideoNameSearchKeywordMaker", "year to int", err)
		iYear = 0
	}
	searchKeyword := title
	if iYear >= 2020 {
		searchKeyword = searchKeyword + " " + year
	}

	return searchKeyword
}

// SearchMatchedVideoFileFromDirs 搜索符合后缀名的视频文件
func SearchMatchedVideoFileFromDirs(l *logrus.Logger, dirs []string) ([]string, error) {

	var fileFullPathList = make([]string, 0)
	for _, dir := range dirs {

		matchedVideoFile, err := SearchMatchedVideoFile(l, dir)
		if err != nil {
			return nil, err
		}

		fileFullPathList = append(fileFullPathList, matchedVideoFile...)
	}

	return fileFullPathList, nil
}

// SearchMatchedVideoFile 搜索符合后缀名的视频文件，现在也会把 BDMV 的文件搜索出来，但是这个并不是一个视频文件，需要在后续特殊处理
func SearchMatchedVideoFile(l *logrus.Logger, dir string) ([]string, error) {

	var fileFullPathList = make([]string, 0)
	pathSep := string(os.PathSeparator)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, curFile := range files {
		fullPath := dir + pathSep + curFile.Name()
		if curFile.IsDir() {
			// 内层的错误就无视了
			oneList, _ := SearchMatchedVideoFile(l, fullPath)
			if oneList != nil {
				fileFullPathList = append(fileFullPathList, oneList...)
			}
		} else {
			// 这里就是文件了
			bok, fakeBDMVVideoFile := FileNameIsBDMV(fullPath)
			if bok == true {
				// 这类文件后续的扫描字幕操作需要额外的处理
				fileFullPathList = append(fileFullPathList, fakeBDMVVideoFile)
				continue
			}
			if IsWantedVideoExtDef(curFile.Name()) == false {
				// 不是期望的视频后缀名则跳过
				continue
			} else {

				// 跳过不符合的文件，比如 MAC OS 下可能有缓存文件，见 #138
				fi, err := curFile.Info()
				if err != nil {
					l.Debugln("SearchMatchedVideoFile, file.Info:", fullPath, err)
					continue
				}
				if fi.Size() == 4096 && strings.HasPrefix(curFile.Name(), "._") == true {
					l.Debugln("SearchMatchedVideoFile file.Size() == 4096 && Prefix Name == ._*", fullPath)
					continue
				}
				fileFullPathList = append(fileFullPathList, fullPath)
			}
		}
	}
	return fileFullPathList, nil
}

// FileNameIsBDMV 是否是 BDMV 蓝光目录，符合返回 true，以及 fakseVideoFPath
func FileNameIsBDMV(id_bdmv_fileFPath string) (bool, string) {
	/*
		这类蓝光视频比较特殊，它没有具体的一个后缀名的视频文件而是由两个文件夹来存储视频数据
		* BDMV
		* CERTIFICATE
		但是不管如何，都需要使用一个文件作为锚点，就选定 CERTIFICATE 中的 id.bdmv 文件
		后续的下载逻辑也需要单独为这个文件进行处理，比如，从这个文件向上一层获取 nfo 文件，
		以及再上一层得到视频文件夹名称等
	*/

	if strings.ToLower(filepath.Base(id_bdmv_fileFPath)) == common.FileBDMV {

		// 这个文件是确认了，那么就需要查看这个文件父级目录是不是 CERTIFICATE 文件夹
		// 且 CERTIFICATE 需要和 BDMV 文件夹都存在
		CERDir := filepath.Dir(id_bdmv_fileFPath)
		BDMVDir := filepath.Join(filepath.Dir(CERDir), "BDMV")
		if IsDir(CERDir) == true && IsDir(BDMVDir) == true {
			return true, filepath.Join(filepath.Dir(CERDir), filepath.Base(filepath.Dir(CERDir))+common.VideoExtMp4)
		}
	}

	return false, ""
}

func SearchTVNfo(l *logrus.Logger, dir string) ([]string, error) {

	var fileFullPathList = make([]string, 0)
	pathSep := string(os.PathSeparator)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, curFile := range files {
		fullPath := dir + pathSep + curFile.Name()
		if curFile.IsDir() {
			// 内层的错误就无视了
			oneList, _ := SearchTVNfo(l, fullPath)
			if oneList != nil {
				fileFullPathList = append(fileFullPathList, oneList...)
			}
		} else {
			// 这里就是文件了
			if strings.ToLower(curFile.Name()) != decode.MetadateTVNfo {
				continue
			} else {

				// 跳过不符合的文件，比如 MAC OS 下可能有缓存文件，见 #138
				fi, err := curFile.Info()
				if err != nil {
					l.Debugln("SearchTVNfo, file.Info:", fullPath, err)
					continue
				}
				if fi.Size() == 4096 && strings.HasPrefix(curFile.Name(), "._") == true {
					l.Debugln("SearchTVNfo file.Size() == 4096 && Prefix Name == ._*", fullPath)
					continue
				}
				fileFullPathList = append(fileFullPathList, fullPath)
			}
		}
	}
	return fileFullPathList, nil
}

// IsWantedVideoExtDef 后缀名是否符合规则
func IsWantedVideoExtDef(fileName string) bool {

	if len(_wantedExtMap) < 1 {
		_defExtMap[common.VideoExtMp4] = common.VideoExtMp4
		_defExtMap[common.VideoExtMkv] = common.VideoExtMkv
		_defExtMap[common.VideoExtRmvb] = common.VideoExtRmvb
		_defExtMap[common.VideoExtIso] = common.VideoExtIso
		_defExtMap[common.VideoExtM2ts] = common.VideoExtM2ts

		_wantedExtMap[common.VideoExtMp4] = common.VideoExtMp4
		_wantedExtMap[common.VideoExtMkv] = common.VideoExtMkv
		_wantedExtMap[common.VideoExtRmvb] = common.VideoExtRmvb
		_wantedExtMap[common.VideoExtIso] = common.VideoExtIso
		_wantedExtMap[common.VideoExtM2ts] = common.VideoExtM2ts

		for _, videoExt := range _customVideoExts {
			_wantedExtMap[videoExt] = videoExt
		}
	}
	fileExt := strings.ToLower(filepath.Ext(fileName))
	_, bFound := _wantedExtMap[fileExt]
	return bFound
}

func GetEpisodeKeyName(season, eps int) string {
	return "S" + strconv.Itoa(season) + "E" + strconv.Itoa(eps)
}

// CopyFile copies a single file from src to dst
func CopyFile(src, dst string) error {
	var err error
	var srcFd *os.File
	var dstFd *os.File
	var srcInfo os.FileInfo

	if srcFd, err = os.Open(src); err != nil {
		return err
	}
	defer func() {
		_ = srcFd.Close()
	}()

	if dstFd, err = os.Create(dst); err != nil {
		return err
	}
	defer func() {
		_ = dstFd.Close()
	}()

	if _, err = io.Copy(dstFd, srcFd); err != nil {
		return err
	}
	if srcInfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcInfo.Mode())
}

// CopyDir copies a whole directory recursively
func CopyDir(src string, dst string) error {
	var err error
	var fds []os.DirEntry
	var srcInfo os.FileInfo

	if srcInfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	if fds, err = os.ReadDir(src); err != nil {
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

// CloseChrome 强行结束没有关闭的 Chrome 进程
func CloseChrome(l *logrus.Logger) {

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
		cmdString = "taskkill /F /im chrome.exe"
		command = exec.Command("cmd.exe", "/c", cmdString)
	}
	if sysType == "darwin" {
		// macOS
		// https://stackoverflow.com/questions/57079120/using-exec-command-in-golang-how-do-i-open-a-new-terminal-and-execute-a-command
		cmdString = `tell application "/Applications/Google Chrome.app" to quit`
		command = exec.Command("osascript", "-s", "h", "-e", cmdString)
	}
	if cmdString == "" || command == nil {
		l.Errorln("CloseChrome OS:", sysType)
		return
	}
	err := command.Run()
	if err != nil {
		l.Warningln("CloseChrome", err)
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
	if sysType == "darwin" {
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

func TimeNumber2Time(inputTimeNumber float64) time.Time {
	newTime := time.Time{}.Add(time.Duration(inputTimeNumber * math.Pow10(9)))
	return newTime
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

func Second2Time(sec int64) time.Time {
	return time.Unix(sec, 0)
}

// ReplaceSpecString 替换特殊的字符
func ReplaceSpecString(inString string, rep string) string {
	return regex_things.RegMatchSpString.ReplaceAllString(inString, rep)
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

// Time2SubTimeString 时间转字幕格式的时间字符串
func Time2SubTimeString(inTime time.Time, timeFormat string) string {
	/*
		这里进行时间转字符串的时候有一点比较特殊
		正常来说输出的格式是类似 15:04:05.00
		那么有个问题，字幕的时间格式是 0:00:12.00， 小时，是个数，除非有跨度到 20 小时的视频，不然小时就应该是个数
		这就需要一个额外的函数去处理这些情况
	*/
	outTimeString := inTime.Format(timeFormat)
	if inTime.Hour() > 9 {
		// 小时，两位数
		return outTimeString
	} else {
		// 小时，一位数
		items := strings.SplitN(outTimeString, ":", -1)
		if len(items) == 3 {

			outTimeString = strings.Replace(outTimeString, items[0], fmt.Sprintf("%d", inTime.Hour()), 1)
			return outTimeString
		}

		return outTimeString
	}
}

// IsEqual 比较 float64
func IsEqual(f1, f2 float64) bool {
	const MIN = 0.000001
	if f1 > f2 {
		return math.Dim(f1, f2) < MIN
	} else {
		return math.Dim(f2, f1) < MIN
	}
}

// ParseTime 解析字幕时间字符串，这里可能小数点后面有 2-4 位
func ParseTime(inTime string) (time.Time, error) {

	parseTime, err := time.Parse(common.TimeFormatPoint2, inTime)
	if err != nil {
		parseTime, err = time.Parse(common.TimeFormatPoint3, inTime)
		if err != nil {
			parseTime, err = time.Parse(common.TimeFormatPoint4, inTime)
		}
	}
	return parseTime, err
}

// GetFileSHA1 获取文件的 SHA1 值
func GetFileSHA1(srcFileFPath string) (string, error) {

	infile, err := os.Open(srcFileFPath)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = infile.Close()
	}()

	h := sha1.New()
	_, err = io.Copy(h, infile)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// WriteFile 写文件
func WriteFile(desFileFPath string, bytes []byte) error {
	var err error
	nowDesPath := desFileFPath
	if filepath.IsAbs(nowDesPath) == false {
		nowDesPath, err = filepath.Abs(nowDesPath)
		if err != nil {
			return err
		}
	}
	// 创建对应的目录
	nowDirPath := filepath.Dir(nowDesPath)
	err = os.MkdirAll(nowDirPath, os.ModePerm)
	if err != nil {
		return err
	}
	file, err := os.Create(nowDesPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	_, err = file.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}

// GetNowTimeString 获取当前的时间，没有秒
func GetNowTimeString() (string, int, int, int) {
	nowTime := time.Now()
	addString := fmt.Sprintf("%d-%d-%d", nowTime.Hour(), nowTime.Minute(), nowTime.Nanosecond())
	return addString, nowTime.Hour(), nowTime.Minute(), nowTime.Nanosecond()
}

// GenerateAccessToken 生成随机的 AccessToken
func GenerateAccessToken() string {
	u4 := uuid.New()
	return u4.String()
}

func UrlJoin(hostUrl, subUrl string) (string, error) {

	u, err := url.Parse(hostUrl)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, subUrl)
	return u.String(), nil
}

// GetFileSHA1String 获取文件的 SHA1 字符串
func GetFileSHA1String(fileFPath string) (string, error) {
	h := sha1.New()

	fp, err := os.Open(fileFPath)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = fp.Close()
	}()

	partAll, err := io.ReadAll(fp)
	if err != nil {
		return "", err
	}

	h.Write(partAll)
	hashBytes := h.Sum(nil)

	return fmt.Sprintf("%x", md5.Sum(hashBytes)), nil
}

var (
	_wantedExtMap    = make(map[string]string) // 人工确认的需要监控的视频后缀名
	_defExtMap       = make(map[string]string) // 内置支持的视频后缀名列表
	_customVideoExts = make([]string, 0)       // 用户额外自定义的视频后缀名列表
)
