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
	"time"
)

func getImdbAndYearMovieXml(movieFilePath string) (common.VideoInfo, error) {

	videoInfo := common.VideoInfo{}
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(movieFilePath); err != nil {
		return videoInfo, err
	}
	for _, t := range doc.FindElements("//IMDB") {
		videoInfo.ImdbId = t.Text()
		break
	}
	for _, t := range doc.FindElements("//ProductionYear") {
		videoInfo.Year = t.Text()
		break
	}
	if videoInfo.ImdbId != "" {
		return videoInfo, nil
	}
	return videoInfo, common.CanNotFindIMDBID
}

func getImdbAndYearNfo(nfoFilePath string) (common.VideoInfo, error) {
	// TODO 新增 TVDB ID 的读取
	imdbInfo := common.VideoInfo{}
	doc := etree.NewDocument()
	// 这里会遇到一个梗，下面的关键词，可能是小写、大写、首字母大写
	// 读取文件转换为全部的小写，然后在解析 xml ？ etree 在转换为小写后，某些类型的文件的内容会崩溃···
	// 所以这里很傻的方式解决
	err := doc.ReadFromFile(nfoFilePath)
	if err != nil {
		return imdbInfo, err
	}
	for _, t := range doc.FindElements("//uniqueid[@type='imdb']") {
		imdbInfo.ImdbId = t.Text()
		break
	}
	for _, t := range doc.FindElements("//uniqueid[@type='Imdb']") {
		imdbInfo.ImdbId = t.Text()
		break
	}
	for _, t := range doc.FindElements("//uniqueid[@type='IMDB']") {
		imdbInfo.ImdbId = t.Text()
		break
	}
	for _, t := range doc.FindElements("./movie/year") {
		imdbInfo.Year = t.Text()
		break
	}
	if imdbInfo.ImdbId != "" {
		return imdbInfo, nil
	}
	return imdbInfo, common.CanNotFindIMDBID
}

func GetImdbInfo(dirPth string) (common.VideoInfo, error) {

	imdbInfo := common.VideoInfo{}
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return imdbInfo, err
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
		if upperName == strings.ToUpper(MetadataFileEmby) {
			movieFilePath = dirPth + pathSep + fi.Name()
			break
		} else if upperName == strings.ToUpper(MetadateTVNfo) {
			// 连续剧的 nfo 文件
			nfoFilePath = dirPth + pathSep + fi.Name()
			break
		} else {
			// 找 *.nfo
			ok := strings.HasSuffix(fi.Name(), suffixNameNfo)
			if ok {
				nfoFilePath = dirPth + pathSep + fi.Name()
			}
		}
	}
	// 根据找到的开始解析
	if movieFilePath == "" && nfoFilePath == "" {
		return imdbInfo, common.NoMetadataFile
	}

	if movieFilePath != "" {
		imdbInfo, err = getImdbAndYearMovieXml(movieFilePath)
		if err != nil {
			GetLogger().Errorln("getImdbAndYearMovieXml error, move on:", err)
		} else {
			return imdbInfo, nil
		}
	}

	if nfoFilePath != "" {
		imdbInfo, err = getImdbAndYearNfo(nfoFilePath)
		if err != nil {
			return imdbInfo, err
		} else {
			return imdbInfo, nil
		}
	}

	return imdbInfo, common.CanNotFindIMDBID
}

//GetVideoInfoFromFileFullPath 从全文件路径推断文件信息
func GetVideoInfoFromFileFullPath(videoFileFullPath string) (*PTN.TorrentInfo, time.Time, error) {

	parse, err := PTN.Parse(filepath.Base(videoFileFullPath))
	if err != nil {
		return nil, time.Time{}, err
	}
	compile, err := regexp.Compile(regFixTitle2)
	if err != nil {
		return nil, time.Time{}, err
	}
	match := compile.ReplaceAllString(parse.Title, "")
	match = strings.TrimRight(match, "")
	parse.Title = match

	fInfo, err := os.Stat(videoFileFullPath)
	if err != nil {
		return nil, time.Time{}, err
	}

	return parse, fInfo.ModTime(), nil
}

// GetSeasonAndEpisodeFromSubFileName 从文件名推断 季 和 集 的信息 Season Episode
func GetSeasonAndEpisodeFromSubFileName(videoFileName string) (bool, int, int, error) {
	// 先进行单个 Episode 的匹配
	// Killing.Eve.S02E01.Do.You.Know.How
	var re = regexp.MustCompile(`(?m)\.S(\d+)E(\d+)\.`)
	matched := re.FindAllStringSubmatch(videoFileName, -1)
	if len(matched) < 1 {
		// Killing.Eve.S02.Do.You.Know.How
		// 看看是不是季度字幕打包
		re = regexp.MustCompile(`(?m)\.S(\d+)\.`)
		matched = re.FindAllStringSubmatch(videoFileName, -1)
		if len(matched) < 1 {
			return false, 0, 0, nil
		}
		season, err := GetNumber2int(matched[0][1])
		if err != nil {
			return false,0, 0, err
		}
		return true, season, 0, nil
	} else {
		// 一集的字幕
		season, err := GetNumber2int(matched[0][1])
		if err != nil {
			return false,0, 0, err
		}
		episode, err := GetNumber2int(matched[0][2])
		if err != nil {
			return false, 0, 0, err
		}

		return false, season, episode, nil
	}
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
	MetadataFileEmby = "movie.xml"
	suffixNameXml    = ".xml"
	suffixNameNfo    = ".nfo"
	MetadateTVNfo    = "tvshow.nfo"
	// 去除特殊字符，仅仅之有中文
	regFixTitle = "[^\u4e00-\u9fa5a-zA-Z0-9\\s]"
	// 去除特殊字符，把特殊字符都写进去
	regFixTitle2 = "[`~!@#$%^&*()+-=|{}';'\\[\\].<>/?~！@#￥%……&*（）——+|{}【】'；”“’。、？]"
	// 获取数字
	regGetNumber = "(?:\\-)?\\d{1,}(?:\\.\\d{1,})?"
)