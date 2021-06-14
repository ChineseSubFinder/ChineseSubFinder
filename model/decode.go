package model

import (
	"errors"
	"github.com/allanpk716/ChineseSubFinder/common"
	"github.com/beevik/etree"
	PTN "github.com/middelink/go-parse-torrent-name"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func getImdbAndYearMovieXml(movieFilePath string) (string, string, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(movieFilePath); err != nil {
		return "", "", err
	}
	imdbId := ""
	for _, t := range doc.FindElements("//IMDB") {
		imdbId = t.Text()
		break
	}
	year := ""
	for _, t := range doc.FindElements("//ProductionYear") {
		year = t.Text()
		break
	}
	if imdbId != "" {
		return imdbId, year, nil
	}
	return "", "", common.CanNotFindIMDBID
}

func getImdbAndYearNfo(nfoFilePath string) (string, string, error) {
	doc := etree.NewDocument()
	// 这里会遇到一个梗，下面的关键词，可能是小写、大写、首字母大写
	// 读取文件转换为全部的小写，然后在解析 xml ？ etree 在转换为小写后，某些类型的文件的内容会崩溃···
	// 所以这里很傻的方式解决
	err := doc.ReadFromFile(nfoFilePath)
	if err != nil {
		return "", "", err
	}
	imdbId := ""
	for _, t := range doc.FindElements("//uniqueid[@type='imdb']") {
		imdbId = t.Text()
		break
	}
	for _, t := range doc.FindElements("//uniqueid[@type='Imdb']") {
		imdbId = t.Text()
		break
	}
	for _, t := range doc.FindElements("//uniqueid[@type='IMDB']") {
		imdbId = t.Text()
		break
	}
	year := ""
	for _, t := range doc.FindElements("./movie/year") {
		year = t.Text()
		break
	}
	if imdbId != "" {
		return imdbId, year, nil
	}
	return "",  "", common.CanNotFindIMDBID
}

func GetImdbIdAndYear(dirPth string) (string, string, error) {
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return "", "", err
	}
	pathSep := string(os.PathSeparator)
	// 优先找 movie.xml 这个是 raddarr 下载的电影会存下来的，可以在 Metadata 设置 Emby
	var movieFilePath = ""
	// 这个是使用 tinyMediaManager 削刮器按 Kodi 来存储的
	var nfoFilePath = ""

	for _, fi := range dir {
		if fi.IsDir() == true {
			continue
		}
		upperName := strings.ToUpper(fi.Name())
		// 找 movie.xml
		if upperName == strings.ToUpper(metadataFileEmby) {
			movieFilePath = dirPth + pathSep + fi.Name()
		}
		// 找 *.nfo
		ok := strings.HasSuffix(fi.Name(), suffixNameNfo)
		if ok {
			nfoFilePath = dirPth + pathSep + fi.Name()
		}
	}
	// 根据找到的开始解析
	if movieFilePath == "" && nfoFilePath == "" {
		return "", "", common.NoMetadataFile
	}

	if movieFilePath != "" {
		outId, outYear, err := getImdbAndYearMovieXml(movieFilePath)
		if err != nil {
			GetLogger().Errorln("getImdbAndYearMovieXml error, move on:", err)
		} else {
			return outId, outYear, nil
		}
	}

	if nfoFilePath != "" {
		outId, outYear, err := getImdbAndYearNfo(nfoFilePath)
		if err != nil {
			return "","", err
		} else {
			return outId, outYear, nil
		}
	}

	return "", "", common.CanNotFindIMDBID
}

//GetVideoInfo 从文件名推断视频文件的信息
func GetVideoInfo(videoFileName string) (*PTN.TorrentInfo, error) {

	parse, err := PTN.Parse(filepath.Base(videoFileName))
	if err != nil {
		return nil, err
	}
	compile, err := regexp.Compile(regFixTitle2)
	if err != nil {
		return nil, err
	}
	match := compile.ReplaceAllString(parse.Title, "")
	match = strings.TrimRight(match, "")
	parse.Title = match
	return parse, nil
}

func GetNumber2Float(input string) (float32, error) {
	compile := regexp.MustCompile(regGetNumber)
	params := compile.FindStringSubmatch(input)
	if len(params) == 0 {
		return 0, errors.New("get number not match")
	}
	fNum, err := strconv.ParseFloat(params[0],32)
	if err != nil {
		return 0, errors.New("get number ParseFloat error")
	}
	return float32(fNum), nil
}

func GetNumber2int(input string) (int, error) {
	compile := regexp.MustCompile(regGetNumber)
	params := compile.FindStringSubmatch(input)
	if len(params) == 0 {
		return 0, errors.New("get number not match")
	}
	fNum, err := strconv.Atoi(params[0])
	if err != nil {
		return 0, errors.New("get number ParseFloat error")
	}
	return fNum, nil
}

const (
	metadataFileEmby = "movie.xml"
	suffixNameXml    = ".xml"
	suffixNameNfo    = ".nfo"
	// 去除特殊字符，仅仅之有中文
	regFixTitle = "[^\u4e00-\u9fa5a-zA-Z0-9\\s]"
	// 去除特殊字符，把特殊字符都写进去
	regFixTitle2 = "[`~!@#$%^&*()+-=|{}';'\\[\\].<>/?~！@#￥%……&*（）——+|{}【】'；”“’。、？]"
	// 获取数字
	regGetNumber = "(?:\\-)?\\d{1,}(?:\\.\\d{1,})?"
)