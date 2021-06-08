package common

import (
	"errors"
	"github.com/beevik/etree"
	PTN "github.com/middelink/go-parse-torrent-name"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func getImdbMovieXml(movieFilePath string) (string, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(movieFilePath); err != nil {
		return "", err
	}
	for _, t := range doc.FindElements("//IMDB") {
		return t.Text(), nil
	}

	return "", CanNotFindIMDBID
}

func getImdbNfo(nfoFilePath string) (string, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(nfoFilePath); err != nil {
		return "", err
	}
	for _, t := range doc.FindElements("//uniqueid[@type='Imdb']") {
		return t.Text(), nil
	}

	return "", CanNotFindIMDBID
}

func GetImdbId(dirPth string) (string ,error) {
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return "", err
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
		return "", NoMetadataFile
	}

	if movieFilePath != "" {
		outId, err := getImdbMovieXml(movieFilePath)
		if err != nil {
			println(err)
		} else {
			return outId, nil
		}
	}

	if nfoFilePath != "" {
		outId, err := getImdbNfo(nfoFilePath)
		if err != nil {
			return "", err
		} else {
			return outId, nil
		}
	}

	return "", CanNotFindIMDBID
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