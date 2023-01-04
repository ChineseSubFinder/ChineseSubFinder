package decode

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"

	"github.com/beevik/etree"
	PTN "github.com/middelink/go-parse-torrent-name"
)

func getVideoNfoInfoFromMovieXml(movieFilePath string) (types.VideoNfoInfo, error) {

	videoInfo := types.VideoNfoInfo{}
	doc := etree.NewDocument()
	doc.ReadSettings.Permissive = true
	if err := doc.ReadFromFile(movieFilePath); err != nil {
		return videoInfo, err
	}
	// --------------------------------------------------
	// IMDB
	for _, t := range doc.FindElements("//imdb") {
		videoInfo.ImdbId = t.Text()
		break
	}
	for _, t := range doc.FindElements("//IMDB") {
		videoInfo.ImdbId = t.Text()
		break
	}
	for _, t := range doc.FindElements("//Imdb") {
		videoInfo.ImdbId = t.Text()
		break
	}
	// --------------------------------------------------
	// TMDB
	for _, t := range doc.FindElements("//tmdb") {
		videoInfo.TmdbId = t.Text()
		break
	}
	for _, t := range doc.FindElements("//TMDB") {
		videoInfo.TmdbId = t.Text()
		break
	}
	for _, t := range doc.FindElements("//Tmdb") {
		videoInfo.TmdbId = t.Text()
		break
	}
	// --------------------------------------------------
	for _, t := range doc.FindElements("//ProductionYear") {
		videoInfo.Year = t.Text()
		break
	}
	//if videoInfo.ImdbId != "" {
	//	return videoInfo, nil
	//}
	videoInfo.IsMovie = true
	return videoInfo, nil
}

func getVideoNfoInfo(nfoFilePath string, rootKey string) (types.VideoNfoInfo, error) {
	imdbInfo := types.VideoNfoInfo{}
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
	// IMDB
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
	// TMDB
	for _, t := range doc.FindElements("./" + rootKey + "/tmdbid") {
		imdbInfo.TmdbId = t.Text()
		break
	}
	for _, t := range doc.FindElements("./" + rootKey + "/tmdb_id") {
		imdbInfo.TmdbId = t.Text()
		break
	}
	for _, t := range doc.FindElements("//uniqueid[@type='tmdb']") {
		imdbInfo.TmdbId = t.Text()
		break
	}
	for _, t := range doc.FindElements("//uniqueid[@type='Tmdb']") {
		imdbInfo.TmdbId = t.Text()
		break
	}
	for _, t := range doc.FindElements("//uniqueid[@type='TMDB']") {
		imdbInfo.TmdbId = t.Text()
		break
	}
	//---------------------------------------------------------------------
	// TVDB
	for _, t := range doc.FindElements("./" + rootKey + "/tvdbid") {
		imdbInfo.TVdbId = t.Text()
		break
	}
	for _, t := range doc.FindElements("./" + rootKey + "/tvdb_id") {
		imdbInfo.TVdbId = t.Text()
		break
	}
	for _, t := range doc.FindElements("//uniqueid[@type='tvdb']") {
		imdbInfo.TVdbId = t.Text()
		break
	}
	for _, t := range doc.FindElements("//uniqueid[@type='Tvdb']") {
		imdbInfo.TVdbId = t.Text()
		break
	}
	for _, t := range doc.FindElements("//uniqueid[@type='TVDB']") {
		imdbInfo.TVdbId = t.Text()
		break
	}
	//---------------------------------------------------------------------
	//Season        int
	//Episode       int
	for _, t := range doc.FindElements("./" + rootKey + "/Season") {
		season, err := strconv.Atoi(t.Text())
		if err != nil {
			continue
		}
		imdbInfo.Season = season
		break
	}
	for _, t := range doc.FindElements("./" + rootKey + "/season") {
		season, err := strconv.Atoi(t.Text())
		if err != nil {
			continue
		}
		imdbInfo.Season = season
		break
	}
	for _, t := range doc.FindElements("./" + rootKey + "/SEASON") {
		season, err := strconv.Atoi(t.Text())
		if err != nil {
			continue
		}
		imdbInfo.Season = season
		break
	}
	for _, t := range doc.FindElements("./" + rootKey + "/Episode") {
		episode, err := strconv.Atoi(t.Text())
		if err != nil {
			continue
		}
		imdbInfo.Episode = episode
		break
	}
	for _, t := range doc.FindElements("./" + rootKey + "/episode") {
		episode, err := strconv.Atoi(t.Text())
		if err != nil {
			continue
		}
		imdbInfo.Episode = episode
		break
	}
	for _, t := range doc.FindElements("./" + rootKey + "/EPISODE") {
		episode, err := strconv.Atoi(t.Text())
		if err != nil {
			continue
		}
		imdbInfo.Episode = episode
		break
	}
	//---------------------------------------------------------------------
	for _, t := range doc.FindElements("./" + rootKey + "/year") {
		imdbInfo.Year = t.Text()
		break
	}
	for _, t := range doc.FindElements("./" + rootKey + "/Year") {
		imdbInfo.Year = t.Text()
		break
	}
	for _, t := range doc.FindElements("./" + rootKey + "/YEAR") {
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
	//if imdbInfo.ImdbId != "" {
	//	return imdbInfo, nil
	//}
	return imdbInfo, nil
}

// GetVideoNfoInfo4Movie 从电影视频文件获取 IMDB info，只能确定拿到 IMDB ID 是靠谱的
func GetVideoNfoInfo4Movie(movieFileFullPath string) (types.VideoNfoInfo, error) {
	videoNfoInfo := types.VideoNfoInfo{}
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
		return videoNfoInfo, err
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
		return videoNfoInfo, common.NoMetadataFile
	}
	// 优先分析 movieName.nfo 文件
	if movieNameNfoFPath != "" {
		videoNfoInfo, err = getVideoNfoInfo(movieNameNfoFPath, "movie")
		if err != nil {
			return videoNfoInfo, err
		}
		videoNfoInfo.IsMovie = true
		return videoNfoInfo, nil
	}

	if nfoFilePath != "" {
		videoNfoInfo, err = getVideoNfoInfo(nfoFilePath, "movie")
		videoNfoInfo.IsMovie = true
		if err != nil {
			return videoNfoInfo, err
		} else {
			return videoNfoInfo, nil
		}
	}

	if movieXmlFPath != "" {
		videoNfoInfo, err = getVideoNfoInfoFromMovieXml(movieXmlFPath)
		videoNfoInfo.IsMovie = true
		if err != nil {
		} else {
			return videoNfoInfo, nil
		}
	}

	videoNfoInfo.IsMovie = true
	return videoNfoInfo, common.NoMetadataFile
}

// GetVideoNfoInfo4SeriesDir 从一个连续剧的根目录获取 IMDB info
func GetVideoNfoInfo4SeriesDir(seriesDir string) (types.VideoNfoInfo, error) {
	imdbInfo := types.VideoNfoInfo{}
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
		return imdbInfo, common.NoMetadataFile
	}

	tmp, err := getVideoNfoInfo(nfoFilePath, "tvshow")
	tmp.IsMovie = false
	return tmp, err
}

// GetVideoNfoInfoFromEpisode 从一集获取这个 Series 的 IMDB info
func GetVideoNfoInfoFromEpisode(oneEpFPath string) (types.VideoNfoInfo, error) {

	// 当前季的路径
	EPdir := filepath.Dir(oneEpFPath)
	// 先判断是否存在 tvshow.nfo
	nfoFilePath := ""
	dir, err := os.ReadDir(EPdir)
	if err != nil {
		return types.VideoNfoInfo{}, err
	}
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

		return GetVideoNfoInfo4SeriesDir(seriesDir)

	} else {

		tmp, err := getVideoNfoInfo(nfoFilePath, "tvshow")
		tmp.IsMovie = false
		return tmp, err
	}
}

// GetVideoNfoInfo4OneSeriesEpisode 获取这一集的 IMDB info，可能会因为没有获取到 IMDB ID 而返回 common.CanNotFindIMDBID 错误，但是 imdbInfo 其他信息是可用的
func GetVideoNfoInfo4OneSeriesEpisode(oneEpFPath string) (types.VideoNfoInfo, error) {

	// 从这一集的视频文件全路径去推算对应的 nfo 文件是否存在
	EPdir := filepath.Dir(oneEpFPath)
	// 与 EP 文件名一致的 nfo 文件名称
	EpNfoFileName := filepath.Base(oneEpFPath)
	EpNfoFileName = strings.ReplaceAll(EpNfoFileName, filepath.Ext(oneEpFPath), suffixNameNfo)
	// 全路径
	EpNfoFPath := filepath.Join(EPdir, EpNfoFileName)

	tmp, err := getVideoNfoInfo(EpNfoFPath, "episodedetails")
	tmp.IsMovie = false
	return tmp, err
}

// GetSeriesDirRootFPath 从一集的绝对路径推断这个连续剧的根目录绝对路径
func GetSeriesDirRootFPath(oneEpFPath string) string {

	oneSeasonDirFPath := filepath.Dir(oneEpFPath)
	oneSeriesDirFPath := filepath.Dir(oneSeasonDirFPath)
	if IsFile(filepath.Join(oneSeriesDirFPath, MetadateTVNfo)) == true {
		return oneSeriesDirFPath
	} else {
		return ""
	}
}

// GetVideoInfoFromFileName 从文件名推断文件信息，这个应该是次要方案，优先还是从 nfo 文件获取这些信息
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

//GetVideoInfoFromFileFullPath 从全文件路径推断文件信息，这个应该是次要方案，优先还是从 nfo 文件获取这些信息
func GetVideoInfoFromFileFullPath(videoFileFullPath string, isMovie bool) (types.VideoNfoInfo, time.Time, error) {

	var err error
	var videoNfoInfo types.VideoNfoInfo
	if isMovie == true {
		videoNfoInfo, err = GetVideoNfoInfo4Movie(videoFileFullPath)
		if err != nil {
			return types.VideoNfoInfo{}, time.Time{}, err
		}

	} else {
		videoNfoInfo, err = GetVideoNfoInfo4OneSeriesEpisode(videoFileFullPath)
		if err != nil {
			return types.VideoNfoInfo{}, time.Time{}, err
		}
	}
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
			return types.VideoNfoInfo{}, time.Time{}, err
		}

		videoNfoInfo.IsMovie = isMovie
		return videoNfoInfo, fInfo.ModTime(), nil

	} else {
		// 再次判断是否是蓝光结构
		// 因为在前面扫描视频的时候，发现特殊的蓝光结构会伪造一个不存在的 xx.mp4 的视频文件过来，这里就需要额外检测一次
		bok, idBDMVFPath, _ := IsFakeBDMVWorked(videoFileFullPath)
		if bok == false {
			return types.VideoNfoInfo{}, time.Time{}, errors.New("GetVideoInfoFromFileFullPath.IsFakeBDMVWorked == false")
		}

		// 获取这个蓝光 ID BDMV 文件的时间
		fInfo, err := os.Stat(idBDMVFPath)
		if err != nil {
			return types.VideoNfoInfo{}, time.Time{}, err
		}

		videoNfoInfo.IsMovie = isMovie
		return videoNfoInfo, fInfo.ModTime(), nil
	}
}

// GetSeasonAndEpisodeFromSubFileName 从文件名推断 季 和 集 的信息 Season Episode，这个应该是次要方案，优先还是从 nfo 文件获取这些信息
func GetSeasonAndEpisodeFromSubFileName(videoFileName string) (bool, int, int, error) {
	upperName := strings.ToUpper(videoFileName)
	// 先进行单个 Episode 的匹配
	// Killing.Eve.S02E01.Do.You.Know.How
	var re = regexp.MustCompile(`(?m)[\.\s]S(\d+).*?E(\d+)[\.\s]`)
	matched := re.FindAllStringSubmatch(upperName, -1)
	if matched == nil || len(matched) < 1 {
		// Killing.Eve.S02.Do.You.Know.How
		// 看看是不是季度字幕打包
		re = regexp.MustCompile(`(?m)[\.\s]S(\d+)[\.\s]`)
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
	idBDMVFPath := filepath.Join(CERDir, common.FileBDMV)

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
	regFixTitle2 = "[~!@#$%^&*:()\\+\\-=|{}';'\\[\\].<>/?~！@#￥%……&*（）——+|{}【】'；”“’。、？]"
	// 获取数字
	regGetNumber = "(?:\\-)?\\d{1,}(?:\\.\\d{1,})?"
)
