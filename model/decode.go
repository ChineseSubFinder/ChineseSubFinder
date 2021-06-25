package model

import (
	"errors"
	"github.com/allanpk716/ChineseSubFinder/common"
	"github.com/beevik/etree"
	PTN "github.com/middelink/go-parse-torrent-name"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func getImdbAndYearMovieXml(movieFilePath string) (common.VideoIMDBInfo, error) {

	videoInfo := common.VideoIMDBInfo{}
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

func getImdbAndYearNfo(nfoFilePath string, rootKey string) (common.VideoIMDBInfo, error) {
	imdbInfo := common.VideoIMDBInfo{}
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
	for _, t := range doc.FindElements("./" + rootKey +"/year") {
		imdbInfo.Year = t.Text()
		break
	}
	for _, t := range doc.FindElements("./" + rootKey + "/releasedate") {
		imdbInfo.ReleaseDate = t.Text()
		break
	}
	for _, t := range doc.FindElements("./" + rootKey + "/premiered") {
		imdbInfo.ReleaseDate = t.Text()
		break
	}
	if imdbInfo.ImdbId != "" {
		return imdbInfo, nil
	}
	return imdbInfo, common.CanNotFindIMDBID
}

func GetImdbInfo4Movie(movieFileFullPath string) (common.VideoIMDBInfo, error) {
	imdbInfo := common.VideoIMDBInfo{}
	// movie 当前的目录
	dirPth := filepath.Dir(movieFileFullPath)
	// 与 movie 文件名一致的 nfo 文件名称
	movieNfoFileName := filepath.Base(movieFileFullPath)
	movieNfoFileName = strings.ReplaceAll(movieNfoFileName, filepath.Ext(movieFileFullPath), suffixNameNfo)
	// movie.xml
	movieXmlFPath := ""
	// movieName.nfo 文件
	movieNameNfoFPath := ""
	// 通用的 *.nfo
	nfoFilePath := ""
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return imdbInfo, err
	}
	for _, fi := range dir {
		if fi.IsDir() == true {
			continue
		}
		upperName := strings.ToLower(fi.Name())
		if upperName == MetadataMovieXml {
			// 找 movie.xml
			movieXmlFPath = path.Join(dirPth, fi.Name())
			break
		} else if upperName == movieNfoFileName {
			// movieName.nfo 文件
			movieNameNfoFPath = path.Join(dirPth, fi.Name())
			break
		} else {
			// 找 *.nfo，很可能是 movie.nfo
			ok := strings.HasSuffix(fi.Name(), suffixNameNfo)
			if ok {
				nfoFilePath = path.Join(dirPth, fi.Name())
			}
		}
	}
	// 根据找到的开始解析
	if movieNameNfoFPath == "" && movieXmlFPath == "" && nfoFilePath == "" {
		return imdbInfo, common.NoMetadataFile
	}
	// 优先分析 movieName.nfo 文件
	if movieNameNfoFPath != "" {
		imdbInfo, err = getImdbAndYearNfo(movieNameNfoFPath, "movie")
		if err != nil {
			return common.VideoIMDBInfo{}, err
		}
		return imdbInfo, nil
	}


	if movieXmlFPath != "" {
		imdbInfo, err = getImdbAndYearMovieXml(movieXmlFPath)
		if err != nil {
			GetLogger().Errorln("getImdbAndYearMovieXml error, move on:", err)
		} else {
			return imdbInfo, nil
		}
	}
	if nfoFilePath != "" {
		imdbInfo, err = getImdbAndYearNfo(nfoFilePath, "movie")
		if err != nil {
			return imdbInfo, err
		} else {
			return imdbInfo, nil
		}
	}

	return imdbInfo, common.CanNotFindIMDBID
}

func GetImdbInfo4SeriesDir(seriesDir string) (common.VideoIMDBInfo, error) {
	imdbInfo := common.VideoIMDBInfo{}
	dir, err := ioutil.ReadDir(seriesDir)
	if err != nil {
		return imdbInfo, err
	}
	nfoFilePath := ""
	for _, fi := range dir {
		if fi.IsDir() == true {
			continue
		}
		upperName := strings.ToUpper(fi.Name())
		if upperName == strings.ToUpper(MetadateTVNfo) {
			// 连续剧的 nfo 文件
			nfoFilePath = path.Join(seriesDir, fi.Name())
			break
		} else {
			// 找 *.nfo
			ok := strings.HasSuffix(fi.Name(), suffixNameNfo)
			if ok {
				nfoFilePath = path.Join(seriesDir, fi.Name())
			}
		}
	}
	// 根据找到的开始解析
	if nfoFilePath == "" {
		return imdbInfo, common.NoMetadataFile
	}
	imdbInfo, err = getImdbAndYearNfo(nfoFilePath, "tvshow")
	if err != nil {
		return common.VideoIMDBInfo{}, err
	}
	return imdbInfo, nil
}

func GetImdbInfo4OneSeriesEpisode(oneEpFPath string) (common.VideoIMDBInfo, error) {

	// 从这一集的视频文件全路径去推算对应的 nfo 文件是否存在
	EPdir := filepath.Dir(oneEpFPath)
	// 与 EP 文件名一致的 nfo 文件名称
	EpNfoFileName := filepath.Base(oneEpFPath)
	EpNfoFileName = strings.ReplaceAll(EpNfoFileName, filepath.Ext(oneEpFPath), suffixNameNfo)
	// 全路径
	EpNfoFPath := path.Join(EPdir, EpNfoFileName)
	//
	imdbInfo := common.VideoIMDBInfo{}
	doc := etree.NewDocument()
	// 这里会遇到一个梗，下面的关键词，可能是小写、大写、首字母大写
	// 读取文件转换为全部的小写，然后在解析 xml ？ etree 在转换为小写后，某些类型的文件的内容会崩溃···
	// 所以这里很傻的方式解决
	err := doc.ReadFromFile(EpNfoFPath)
	if err != nil {
		return imdbInfo, err
	}
	for _, t := range doc.FindElements("./episodedetails/aired") {
		imdbInfo.ReleaseDate = t.Text()
		break
	}
	for _, t := range doc.FindElements("./episodedetails/premiered") {
		imdbInfo.ReleaseDate = t.Text()
		break
	}
	if imdbInfo.ReleaseDate != "" {
		return imdbInfo, nil
	}
	return imdbInfo, common.CanNotFindEpAiredTime
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
	upperName := strings.ToUpper(videoFileName)
	// 先进行单个 Episode 的匹配
	// Killing.Eve.S02E01.Do.You.Know.How
	var re = regexp.MustCompile(`(?m)\.S(\d+)E(\d+)\.`)
	matched := re.FindAllStringSubmatch(upperName, -1)
	if len(matched) < 1 {
		// Killing.Eve.S02.Do.You.Know.How
		// 看看是不是季度字幕打包
		re = regexp.MustCompile(`(?m)\.S(\d+)\.`)
		matched = re.FindAllStringSubmatch(upperName, -1)
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
	MetadataMovieXml = "movie.xml"
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