package decode

import (
	"errors"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	common2 "github.com/allanpk716/ChineseSubFinder/internal/types/common"
	"github.com/beevik/etree"
	PTN "github.com/middelink/go-parse-torrent-name"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func getImdbAndYearMovieXml(movieFilePath string) (types.VideoIMDBInfo, error) {

	videoInfo := types.VideoIMDBInfo{}
	doc := etree.NewDocument()
	doc.ReadSettings.Permissive = true
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
	return videoInfo, common2.CanNotFindIMDBID
}

func getImdbAndYearNfo(nfoFilePath string, rootKey string) (types.VideoIMDBInfo, error) {
	imdbInfo := types.VideoIMDBInfo{}
	doc := etree.NewDocument()
	doc.ReadSettings.Permissive = true
	// 这里会遇到一个梗，下面的关键词，可能是小写、大写、首字母大写
	// 读取文件转换为全部的小写，然后在解析 xml ？ etree 在转换为小写后，某些类型的文件的内容会崩溃···
	// 所以这里很傻的方式解决
	err := doc.ReadFromFile(nfoFilePath)
	if err != nil {
		return imdbInfo, err
	}
	for _, t := range doc.FindElements("./" + rootKey + "/title") {
		imdbInfo.Title = t.Text()
		break
	}
	//---------------------------------------------------------------------
	for _, t := range doc.FindElements("./" + rootKey + "/imdbid") {
		imdbInfo.ImdbId = t.Text()
		break
	}
	for _, t := range doc.FindElements("./" + rootKey + "/imdb_id") {
		imdbInfo.ImdbId = t.Text()
		break
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
	//---------------------------------------------------------------------
	for _, t := range doc.FindElements("./" + rootKey + "/year") {
		imdbInfo.Year = t.Text()
		break
	}
	//---------------------------------------------------------------------
	for _, t := range doc.FindElements("./" + rootKey + "/releasedate") {
		imdbInfo.ReleaseDate = t.Text()
		break
	}
	for _, t := range doc.FindElements("./" + rootKey + "/aired") {
		imdbInfo.ReleaseDate = t.Text()
		break
	}
	//---------------------------------------------------------------------
	for _, t := range doc.FindElements("./" + rootKey + "/premiered") {
		imdbInfo.ReleaseDate = t.Text()
		break
	}
	if imdbInfo.ImdbId != "" {
		return imdbInfo, nil
	}
	return imdbInfo, common2.CanNotFindIMDBID
}

// GetImdbInfo4Movie 从电影视频文件获取 IMDB info，只能确定拿到 IMDB ID 是靠谱的
func GetImdbInfo4Movie(movieFileFullPath string) (types.VideoIMDBInfo, error) {
	imdbInfo := types.VideoIMDBInfo{}
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
	dir, err := os.ReadDir(dirPth)
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
			movieXmlFPath = filepath.Join(dirPth, fi.Name())

		} else if upperName == movieNfoFileName {
			// movieName.nfo 文件
			movieNameNfoFPath = filepath.Join(dirPth, fi.Name())

		} else {
			// 找 *.nfo，很可能是 movie.nfo
			ok := strings.HasSuffix(fi.Name(), suffixNameNfo)
			if ok {
				nfoFilePath = filepath.Join(dirPth, fi.Name())
			}
		}
	}
	// 根据找到的开始解析
	if movieNameNfoFPath == "" && movieXmlFPath == "" && nfoFilePath == "" {
		return imdbInfo, common2.NoMetadataFile
	}
	// 优先分析 movieName.nfo 文件
	if movieNameNfoFPath != "" {
		imdbInfo, err = getImdbAndYearNfo(movieNameNfoFPath, "movie")
		if err != nil {
			return imdbInfo, err
		}
		return imdbInfo, nil
	}

	if movieXmlFPath != "" {
		imdbInfo, err = getImdbAndYearMovieXml(movieXmlFPath)
		if err != nil {
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

	return imdbInfo, common2.CanNotFindIMDBID
}

// GetImdbInfo4SeriesDir 从一个连续剧的根目录获取 IMDB info
func GetImdbInfo4SeriesDir(seriesDir string) (types.VideoIMDBInfo, error) {
	imdbInfo := types.VideoIMDBInfo{}
	dir, err := os.ReadDir(seriesDir)
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
			nfoFilePath = filepath.Join(seriesDir, fi.Name())
			break
		} else {
			// 找 *.nfo
			ok := strings.HasSuffix(fi.Name(), suffixNameNfo)
			if ok {
				nfoFilePath = filepath.Join(seriesDir, fi.Name())
			}
		}
	}
	// 根据找到的开始解析
	if nfoFilePath == "" {
		return imdbInfo, common2.NoMetadataFile
	}
	imdbInfo, err = getImdbAndYearNfo(nfoFilePath, "tvshow")
	if err != nil {
		return imdbInfo, err
	}
	return imdbInfo, nil
}

// GetSeriesSeasonImdbInfoFromEpisode 从一集获取这个 Series 的 IMDB info
func GetSeriesSeasonImdbInfoFromEpisode(oneEpFPath string) (types.VideoIMDBInfo, error) {

	var err error
	// 当前季的路径
	EPdir := filepath.Dir(oneEpFPath)
	// 先判断是否存在 tvshow.nfo
	nfoFilePath := ""
	dir, err := os.ReadDir(EPdir)
	for _, fi := range dir {
		if fi.IsDir() == true {
			continue
		}
		upperName := strings.ToUpper(fi.Name())
		if upperName == strings.ToUpper(MetadateTVNfo) {
			// 连续剧的 nfo 文件
			nfoFilePath = filepath.Join(EPdir, fi.Name())
			break
		}
	}
	if nfoFilePath == "" {

		// 没有找到，那么就向上一级再次找
		seasonDir := filepath.Base(EPdir)
		seriesDir := EPdir[:len(EPdir)-len(seasonDir)]

		return GetImdbInfo4SeriesDir(seriesDir)

	} else {
		var imdbInfo types.VideoIMDBInfo
		imdbInfo, err = getImdbAndYearNfo(nfoFilePath, "tvshow")
		if err != nil {
			return imdbInfo, err
		}
		return imdbInfo, nil
	}
}

// GetImdbInfo4OneSeriesEpisode 获取这一集的 IMDB info，可能会因为没有获取到 IMDB ID 而返回 common.CanNotFindIMDBID 错误，但是 imdbInfo 其他信息是可用的
func GetImdbInfo4OneSeriesEpisode(oneEpFPath string) (types.VideoIMDBInfo, error) {

	// 从这一集的视频文件全路径去推算对应的 nfo 文件是否存在
	EPdir := filepath.Dir(oneEpFPath)
	// 与 EP 文件名一致的 nfo 文件名称
	EpNfoFileName := filepath.Base(oneEpFPath)
	EpNfoFileName = strings.ReplaceAll(EpNfoFileName, filepath.Ext(oneEpFPath), suffixNameNfo)
	// 全路径
	EpNfoFPath := filepath.Join(EPdir, EpNfoFileName)

	imdbInfo, err := getImdbAndYearNfo(EpNfoFPath, "episodedetails")
	if err != nil {
		return imdbInfo, err
	}

	return imdbInfo, nil
}

// GetVideoInfoFromFileName 从文件名推断文件信息
func GetVideoInfoFromFileName(fileName string) (*PTN.TorrentInfo, error) {

	parse, err := PTN.Parse(fileName)
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

	/*
		这里有个特殊情况，如果是某一种蓝光的文件结构，不是一个单一的视频文件
		* 失控玩家 (2021)
			* BDMV
			* CERTIFICATE
				* id.bdmv
		大致是这样的目录结构，两个文件夹，下面按个文件夹中一定有这个文件 id.bdmv
		那么，在前期的扫描视频的阶段，会把这样的蓝光视频给伪造一个假的不存在的视频传入进来
		失控玩家 (2021).mp4 比如这个
		然后需要 check 这个文件是否存在：
			1. 如果 check 这个文件存在，那么就是之前的逻辑
			2. 如果是这个情况肯定是不存在的，那么就要判断是否有这文件结构是否符合这种蓝光结构

	*/
	if IsFile(videoFileFullPath) == true {
		// 常见的视频情况
		fInfo, err := os.Stat(videoFileFullPath)
		if err != nil {
			return nil, time.Time{}, err
		}

		return parse, fInfo.ModTime(), nil
	} else {
		// 再次判断是否是蓝光结构
		// 因为在前面扫描视频的时候，发现特殊的蓝光结构会伪造一个不存在的 xx.mp4 的视频文件过来，这里就需要额外检测一次
		bok, idBDMVFPath, _ := IsFakeBDMVWorked(videoFileFullPath)
		if bok == false {
			return nil, time.Time{}, errors.New("GetVideoInfoFromFileFullPath.IsFakeBDMVWorked == false")
		}

		// 获取这个蓝光 ID BDMV 文件的时间
		fInfo, err := os.Stat(idBDMVFPath)
		if err != nil {
			return nil, time.Time{}, err
		}
		return parse, fInfo.ModTime(), nil
	}
}

// GetSeasonAndEpisodeFromSubFileName 从文件名推断 季 和 集 的信息 Season Episode
func GetSeasonAndEpisodeFromSubFileName(videoFileName string) (bool, int, int, error) {
	upperName := strings.ToUpper(videoFileName)
	// 先进行单个 Episode 的匹配
	// Killing.Eve.S02E01.Do.You.Know.How
	var re = regexp.MustCompile(`(?m)\.S(\d+)E(\d+)\.`)
	matched := re.FindAllStringSubmatch(upperName, -1)
	if matched == nil || len(matched) < 1 {
		// Killing.Eve.S02.Do.You.Know.How
		// 看看是不是季度字幕打包
		re = regexp.MustCompile(`(?m)\.S(\d+)\.`)
		matched = re.FindAllStringSubmatch(upperName, -1)
		if matched == nil || len(matched) < 1 {
			return false, 0, 0, nil
		}
		season, err := GetNumber2int(matched[0][1])
		if err != nil {
			return false, 0, 0, err
		}
		return true, season, 0, nil
	} else {
		// 一集的字幕
		season, err := GetNumber2int(matched[0][1])
		if err != nil {
			return false, 0, 0, err
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
	if params == nil || len(params) == 0 {
		return 0, errors.New("get number not match")
	}
	fNum, err := strconv.ParseFloat(params[0], 32)
	if err != nil {
		return 0, errors.New("get number ParseFloat error")
	}
	return float32(fNum), nil
}

func GetNumber2int(input string) (int, error) {
	compile := regexp.MustCompile(regGetNumber)
	params := compile.FindStringSubmatch(input)
	if params == nil || len(params) == 0 {
		return 0, errors.New("get number not match")
	}
	fNum, err := strconv.Atoi(params[0])
	if err != nil {
		return 0, errors.New("get number ParseFloat error")
	}
	return fNum, nil
}

// IsFile 存在且是文件
func IsFile(filePath string) bool {
	s, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	return !s.IsDir()
}

// IsDir 存在且是文件夹
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// IsFakeBDMVWorked 传入的是伪造的不存在的蓝光结构的视频全路径，如果是就返回 true 和 id.bdmv 的绝对路径 和 STREAM 绝对路径
func IsFakeBDMVWorked(fakseVideFPath string) (bool, string, string) {

	rootDir := filepath.Dir(fakseVideFPath)

	CERDir := filepath.Join(rootDir, "CERTIFICATE")
	BDMVDir := filepath.Join(rootDir, "BDMV")
	STREAMDir := filepath.Join(BDMVDir, "STREAM")
	idBDMVFPath := filepath.Join(CERDir, common2.FileBDMV)

	if IsDir(CERDir) == true && IsDir(BDMVDir) == true && IsFile(idBDMVFPath) == true {
		return true, idBDMVFPath, STREAMDir
	}

	return false, "", ""
}

const (
	MetadataMovieXml = "movie.xml"
	suffixNameXml    = ".xml"
	suffixNameNfo    = ".nfo"
	MetadateTVNfo    = "tvshow.nfo"
	// 去除特殊字符，仅仅之有中文
	regFixTitle = "[^\u4e00-\u9fa5a-zA-Z0-9\\s]"
	// 去除特殊字符，把特殊字符都写进去
	regFixTitle2 = "[~!@#$%^&*()\\+\\-=|{}';'\\[\\].<>/?~！@#￥%……&*（）——+|{}【】'；”“’。、？]"
	// 获取数字
	regGetNumber = "(?:\\-)?\\d{1,}(?:\\.\\d{1,})?"
)
