package series_helper

import (
	"path/filepath"
	"strconv"
	"time"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/media_info_dealers"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/search"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/ifaces"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/emby"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/series"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/supplier"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_parser/ass"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_parser/srt"

	"github.com/ChineseSubFinder/ChineseSubFinder/internal/models"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/decode"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/imdb_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_parser_hub"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/jinzhu/now"
	"github.com/sirupsen/logrus"
)

func readSeriesInfo(dealers *media_info_dealers.Dealers, seriesDir string, need2AnalyzeSub bool) (*series.SeriesInfo, map[string][]series.SubInfo, error) {

	seriesInfo, err := GetSeriesInfoFromDir(dealers, seriesDir)
	if err != nil {
		return nil, nil, err
	}
	seriesInfo.NeedDlSeasonDict = make(map[int]int)
	seriesInfo.NeedDlEpsKeyList = make(map[string]series.EpisodeInfo)

	// 字幕字典 S01E01 - []SubInfo
	SubDict := make(map[string][]series.SubInfo)

	if need2AnalyzeSub == false {
		return seriesInfo, SubDict, nil
	}

	subParserHub := sub_parser_hub.NewSubParserHub(dealers.Logger, ass.NewParser(dealers.Logger), srt.NewParser(dealers.Logger))
	// 先搜索这个目录下，所有符合条件的视频
	matchedVideoFile, err := search.MatchedVideoFile(dealers.Logger, seriesDir)
	if err != nil {
		return nil, nil, err
	}
	// 然后再从这个视频找到对用匹配的字幕
	for _, oneVideoFPath := range matchedVideoFile {

		subFiles, err := sub_helper.SearchMatchedSubFileByOneVideo(dealers.Logger, oneVideoFPath)
		if err != nil {
			return nil, nil, err
		}
		epsVideoNfoInfo, err := decode.GetVideoNfoInfo4OneSeriesEpisode(oneVideoFPath)
		if err != nil {
			dealers.Logger.Errorln(err)
			continue
		}

		for _, subFile := range subFiles {

			bFind, subParserFileInfo, err := subParserHub.DetermineFileTypeFromFile(subFile)
			if err != nil {
				dealers.Logger.Errorln("DetermineFileTypeFromFile", subFile, err)
				continue
			}
			if bFind == false {
				dealers.Logger.Warnln("DetermineFileTypeFromFile", subFile, "not support SubType")
				continue
			}
			// 判断这个字幕是否包含中文
			if subParserHub.IsSubHasChinese(subParserFileInfo) == false {
				continue
			}
			epsKey := pkg.GetEpisodeKeyName(epsVideoNfoInfo.Season, epsVideoNfoInfo.Episode)
			oneFileSubInfo := series.SubInfo{
				Title:        epsVideoNfoInfo.Title,
				Season:       epsVideoNfoInfo.Season,
				Episode:      epsVideoNfoInfo.Episode,
				Language:     subParserFileInfo.Lang,
				Dir:          filepath.Dir(subFile),
				FileFullPath: subFile,
			}
			_, ok := SubDict[epsKey]
			if ok == false {
				// 初始化
				SubDict[epsKey] = make([]series.SubInfo, 0)
			}
			SubDict[epsKey] = append(SubDict[epsKey], oneFileSubInfo)
		}
	}

	return seriesInfo, SubDict, nil
}

// ReadSeriesInfoFromDir 读取剧集的信息，只有那些 Eps 需要下载字幕的 NeedDlEpsKeyList
func ReadSeriesInfoFromDir(dealers *media_info_dealers.Dealers,
	seriesDir string,
	ExpirationTime int,
	forcedScanAndDownloadSub bool,
	need2AnalyzeSub bool,
	epsMap ...map[int][]int) (*series.SeriesInfo, error) {

	seriesInfo, SubDict, err := readSeriesInfo(dealers, seriesDir, need2AnalyzeSub)
	if err != nil {
		return nil, err
	}
	// 搜索所有的视频
	videoFiles, err := search.MatchedVideoFile(dealers.Logger, seriesDir)
	if err != nil {
		return nil, err
	}
	// 视频字典 S01E01 - EpisodeInfo
	EpisodeDict := make(map[string]series.EpisodeInfo)
	for _, videoFile := range videoFiles {
		getEpsInfoAndSubDic(dealers.Logger, videoFile, EpisodeDict, SubDict, epsMap...)
	}

	for _, episodeInfo := range EpisodeDict {
		seriesInfo.EpList = append(seriesInfo.EpList, episodeInfo)
		seriesInfo.SeasonDict[episodeInfo.Season] = episodeInfo.Season
	}

	seriesInfo.NeedDlEpsKeyList, seriesInfo.NeedDlSeasonDict = whichSeasonEpsNeedDownloadSub(dealers.Logger, seriesInfo, ExpirationTime, forcedScanAndDownloadSub)

	return seriesInfo, nil
}

// ReadSeriesInfoFromEmby 将 Emby API 读取到的数据进行转换到通用的结构中，需要填充那些剧集需要下载，这样要的是一个连续剧的，不是所有的传入(只有那些 Eps 需要下载字幕的 NeedDlEpsKeyList)
func ReadSeriesInfoFromEmby(dealers *media_info_dealers.Dealers, seriesDir string, seriesVideoList []emby.EmbyMixInfo, ExpirationTime int, forcedScanAndDownloadSub bool, need2AnalyzeSub bool) (*series.SeriesInfo, error) {

	seriesInfo, SubDict, err := readSeriesInfo(dealers, seriesDir, need2AnalyzeSub)
	if err != nil {
		return nil, err
	}

	EpisodeDict := make(map[string]series.EpisodeInfo)
	for _, info := range seriesVideoList {
		getEpsInfoAndSubDic(dealers.Logger, info.PhysicalVideoFileFullPath, EpisodeDict, SubDict)
	}

	for _, episodeInfo := range EpisodeDict {
		seriesInfo.EpList = append(seriesInfo.EpList, episodeInfo)
		seriesInfo.SeasonDict[episodeInfo.Season] = episodeInfo.Season
	}

	seriesInfo.NeedDlEpsKeyList, seriesInfo.NeedDlSeasonDict = whichSeasonEpsNeedDownloadSub(dealers.Logger, seriesInfo, ExpirationTime, forcedScanAndDownloadSub)

	return seriesInfo, nil
}

// SkipChineseSeries 跳过中文连续剧
func SkipChineseSeries(dealers *media_info_dealers.Dealers, seriesRootPath string) (bool, *models.IMDBInfo, error) {

	imdbInfo, err := decode.GetVideoNfoInfo4SeriesDir(seriesRootPath)
	if err != nil {
		return false, nil, err
	}

	isChineseVideo, t, err := imdb_helper.IsChineseVideo(dealers, imdbInfo)
	if err != nil {
		return false, nil, err
	}
	if isChineseVideo == true {
		dealers.Logger.Infoln("Skip", filepath.Base(seriesRootPath), "Sub Download, because series is Chinese")
		return true, t, nil
	} else {
		return false, t, nil
	}
}

// DownloadSubtitleInAllSiteByOneSeries 一部连续剧，在所有的网站，下载相应的字幕
func DownloadSubtitleInAllSiteByOneSeries(logger *logrus.Logger, Suppliers []ifaces.ISupplier, seriesInfo *series.SeriesInfo, i int64) []supplier.SubInfo {

	defer func() {
		logger.Infoln(common.QueueName, i, "DlSub End", seriesInfo.DirPath)
		logger.Infoln("------------------------------------------")
	}()
	logger.Infoln(common.QueueName, i, "DlSub Start", seriesInfo.DirPath)
	logger.Infoln(common.QueueName, i, "IMDB ID:", seriesInfo.ImdbId, "NeedDownloadSubs:", len(seriesInfo.NeedDlEpsKeyList))
	var outSUbInfos = make([]supplier.SubInfo, 0)
	if len(seriesInfo.NeedDlEpsKeyList) < 1 {
		return outSUbInfos
	}
	for key := range seriesInfo.NeedDlEpsKeyList {
		logger.Infoln(common.QueueName, i, "NeedDownloadEps", "-", key)
	}

	for _, oneSupplier := range Suppliers {

		oneSupplierFunc := func() {
			defer func() {
				logger.Infoln(common.QueueName, i, oneSupplier.GetSupplierName(), "End")
				logger.Infoln("------------------------------------------")
			}()

			var subInfos []supplier.SubInfo
			logger.Infoln("------------------------------------------")
			logger.Infoln(common.QueueName, i, oneSupplier.GetSupplierName(), "Start...")

			if oneSupplier.OverDailyDownloadLimit() == true {
				logger.Infoln(common.QueueName, i, oneSupplier.GetSupplierName(), "Over Daily Download Limit")
				return
			}

			// 一次性把这一部连续剧的所有字幕下载完
			subInfos, err := oneSupplier.GetSubListFromFile4Series(seriesInfo)
			if err != nil {
				logger.Errorln(common.QueueName, i, oneSupplier.GetSupplierName(), "GetSubListFromFile4Series", err)
				return
			}
			// 把后缀名给改好
			sub_helper.ChangeVideoExt2SubExt(subInfos)

			outSUbInfos = append(outSUbInfos, subInfos...)
		}

		oneSupplierFunc()
	}

	return outSUbInfos
}

// GetSeriesListFromDirs 获取这个目录下的所有文件夹名称，默认为一个连续剧的目录的List
func GetSeriesListFromDirs(logger *logrus.Logger, dirs []string) (*treemap.Map, error) {

	defer func() {
		logger.Infoln("GetSeriesListFromDirs End")
		logger.Infoln("------------------------------------------")
	}()

	logger.Infoln("------------------------------------------")
	logger.Infoln("GetSeriesListFromDirs Start...")

	var fileFullPathMap = treemap.NewWithStringComparator()
	for _, dir := range dirs {

		seriesList, err := GetSeriesList(logger, dir)
		if err != nil {
			return nil, err
		}

		value, found := fileFullPathMap.Get(dir)
		if found == false {
			fileFullPathMap.Put(dir, seriesList)
		} else {
			value = append(value.([]string), seriesList...)
			fileFullPathMap.Put(dir, value)
		}
	}

	return fileFullPathMap, nil
}

// GetSeriesList 获取这个目录下的所有文件夹名称，默认为一个连续剧的目录的List
func GetSeriesList(log *logrus.Logger, dir string) ([]string, error) {

	// 需要把所有 tvshow.nfo 搜索出来，那么这些文件对应的目录就是目标连续剧的目录
	tvNFOs, err := search.TVNfo(log, dir)
	if err != nil {
		return nil, err
	}
	var seriesDirList = make([]string, 0)

	for _, tvNfo := range tvNFOs {
		seriesDirList = append(seriesDirList, filepath.Dir(tvNfo))
	}

	return seriesDirList, err
}

// whichSeasonEpsNeedDownloadSub 有那些 Eps 需要下载的，按 SxEx 反回 epsKey
func whichSeasonEpsNeedDownloadSub(logger *logrus.Logger, seriesInfo *series.SeriesInfo, ExpirationTime int, forcedScanAndDownloadSub bool) (map[string]series.EpisodeInfo, map[int]int) {
	var needDlSubEpsList = make(map[string]series.EpisodeInfo, 0)
	var needDlSeasonList = make(map[int]int, 0)
	currentTime := time.Now()
	// 直接强制所有视频都下载字幕
	if forcedScanAndDownloadSub == true {
		for _, epsInfo := range seriesInfo.EpList {
			// 添加
			epsKey := pkg.GetEpisodeKeyName(epsInfo.Season, epsInfo.Episode)
			needDlSubEpsList[epsKey] = epsInfo
			needDlSeasonList[epsInfo.Season] = epsInfo.Season
		}

		return needDlSubEpsList, needDlSeasonList
	}

	for _, epsInfo := range seriesInfo.EpList {
		// 如果没有字幕，则加入下载列表
		// 如果每一集的播出时间能够读取到，那么就以这个完后推算 3个月
		// 如果读取不到 Aired Time 那么，这一集下载后的 ModifyTime 3个月天内，都进行字幕的下载
		var err error
		var baseTime time.Time
		if epsInfo.AiredTime != "" {
			baseTime, err = now.Parse(epsInfo.AiredTime)
			if err != nil {
				logger.Errorln("SeriesInfo parse AiredTime", epsInfo.Title, epsInfo.Season, epsInfo.Episode, err)
				baseTime = epsInfo.ModifyTime
			}
		} else {
			baseTime = epsInfo.ModifyTime
		}

		if len(epsInfo.SubAlreadyDownloadedList) < 1 || baseTime.AddDate(0, 0, ExpirationTime).After(currentTime) == true {
			// 添加
			epsKey := pkg.GetEpisodeKeyName(epsInfo.Season, epsInfo.Episode)
			needDlSubEpsList[epsKey] = epsInfo
			needDlSeasonList[epsInfo.Season] = epsInfo.Season
		} else {
			if len(epsInfo.SubAlreadyDownloadedList) > 0 {
				logger.Infoln("Skip because find sub file and downloaded or aired over 3 months,", epsInfo.Title, epsInfo.Season, epsInfo.Episode)
			} else if baseTime.AddDate(0, 0, ExpirationTime).After(currentTime) == false {
				logger.Infoln("Skip because 3 months pass,", epsInfo.Title, epsInfo.Season, epsInfo.Episode)
			}
		}
	}
	return needDlSubEpsList, needDlSeasonList
}

func GetSeriesInfoFromDir(dealers *media_info_dealers.Dealers, seriesDir string) (*series.SeriesInfo, error) {
	seriesInfo := series.SeriesInfo{}
	// 只考虑 IMDB 去查询，文件名目前发现可能会跟电影重复，导致很麻烦，本来也有前置要求要削刮器处理的
	videoInfo, err := decode.GetVideoNfoInfo4SeriesDir(seriesDir)
	if err != nil {
		return nil, err
	}

	imdbInfo, err := imdb_helper.GetIMDBInfoFromVideoNfoInfo(dealers, videoInfo)
	if err != nil {
		return nil, err
	}

	// 使用 IMDB ID 得到通用的剧集名称
	// 以 IMDB 的信息为准
	if imdbInfo != nil {

		if imdbInfo.Name != "" {
			seriesInfo.Name = imdbInfo.Name
		} else if videoInfo.Title != "" {
			seriesInfo.Name = videoInfo.Title
		} else {
			seriesInfo.Name = filepath.Base(seriesDir)
		}
		seriesInfo.ImdbId = imdbInfo.IMDBID
		seriesInfo.Year = imdbInfo.Year
	} else {
		if videoInfo.Title != "" {
			seriesInfo.Name = videoInfo.Title
		} else {
			seriesInfo.Name = filepath.Base(seriesDir)
		}
		seriesInfo.ImdbId = videoInfo.ImdbId
		iYear, err := strconv.Atoi(videoInfo.Year)
		if err != nil {
			// 不是必须的
			seriesInfo.Year = 0
			dealers.Logger.Warnln("ReadSeriesInfoFromDir.GetVideoNfoInfo4SeriesDir.strconv.Atoi", seriesDir, err)
		} else {
			seriesInfo.Year = iYear
		}
	}

	seriesInfo.ReleaseDate = videoInfo.ReleaseDate
	seriesInfo.DirPath = seriesDir
	seriesInfo.EpList = make([]series.EpisodeInfo, 0)
	seriesInfo.SeasonDict = make(map[int]int)
	return &seriesInfo, nil
}

func getEpsInfoAndSubDic(logger *logrus.Logger,
	videoFile string,
	EpisodeDict map[string]series.EpisodeInfo,
	SubDict map[string][]series.SubInfo,
	epsMap ...map[int][]int) {
	// 正常来说，一集只有一个格式的视频，也就是 S01E01 只有一个，如果有多个则会只保存第一个
	episodeInfo, modifyTime, err := decode.GetVideoInfoFromFileFullPath(videoFile, false)
	if err != nil {
		logger.Errorln("model.GetVideoInfoFromFileFullPath", videoFile, err)
		return
	}

	if len(epsMap) > 0 {
		// 如果这个视频不在需要下载的 Eps 列表中，那么就跳过后续的逻辑
		epsList, ok := epsMap[0][episodeInfo.Season]
		if ok == false {
			return
		}
		found := false
		for _, oneEpsID := range epsList {
			if oneEpsID == episodeInfo.Episode {
				// 在需要下载的 Eps 列表中
				found = true
				break
			}
		}
		if found == false {
			return
		}
	}

	epsKey := pkg.GetEpisodeKeyName(episodeInfo.Season, episodeInfo.Episode)
	_, ok := EpisodeDict[epsKey]
	if ok == false {
		// 初始化
		oneFileEpInfo := series.EpisodeInfo{
			Title:        episodeInfo.Title,
			Season:       episodeInfo.Season,
			Episode:      episodeInfo.Episode,
			Dir:          filepath.Dir(videoFile),
			FileFullPath: videoFile,
			ModifyTime:   modifyTime,
			AiredTime:    episodeInfo.ReleaseDate,
		}
		// 需要匹配同级目录下的字幕
		oneFileEpInfo.SubAlreadyDownloadedList = make([]series.SubInfo, 0)
		for _, subInfo := range SubDict[epsKey] {
			if subInfo.Dir == oneFileEpInfo.Dir {
				oneFileEpInfo.SubAlreadyDownloadedList = append(oneFileEpInfo.SubAlreadyDownloadedList, subInfo)
			}
		}
		EpisodeDict[epsKey] = oneFileEpInfo
	} else {
		// 存在则跳过
		return
	}
	return
}
