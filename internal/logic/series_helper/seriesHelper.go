package series_helper

import (
	"github.com/StalkR/imdb"
	"github.com/allanpk716/ChineseSubFinder/internal/common"
	"github.com/allanpk716/ChineseSubFinder/internal/ifaces"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/ass"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/srt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/decode"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/imdb_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	"github.com/allanpk716/ChineseSubFinder/internal/types/emby"
	"github.com/allanpk716/ChineseSubFinder/internal/types/series"
	"github.com/allanpk716/ChineseSubFinder/internal/types/supplier"
	"github.com/jinzhu/now"
	"io/ioutil"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ReadSeriesInfoFromDir 读取剧集的信息，只有那些 Eps 需要下载字幕的 NeedDlEpsKeyList
func ReadSeriesInfoFromDir(seriesDir string, imdbInfo *imdb.Title) (*series.SeriesInfo, error) {

	subParserHub := sub_helper.NewSubParserHub(ass.NewParser(), srt.NewParser())

	seriesInfo, err := getSeriesInfoFromDir(seriesDir, imdbInfo)
	if err != nil {
		return nil, err
	}
	// 搜索所有的视频
	videoFiles, err := pkg.SearchMatchedVideoFile(seriesDir)
	if err != nil {
		return nil, err
	}
	// 搜索所有的字幕
	subFiles, err := sub_helper.SearchMatchedSubFile(seriesDir)
	if err != nil {
		return nil, err
	}
	// 字幕字典 S01E01 - []SubInfo
	SubDict := make(map[string][]series.SubInfo)
	for _, subFile := range subFiles {
		// 判断这个字幕是否包含中文
		if subParserHub.IsSubHasChinese(subFile) == false {
			continue
		}
		info, _, err := decode.GetVideoInfoFromFileFullPath(subFile)
		if err != nil {
			log_helper.GetLogger().Errorln(err)
			continue
		}
		subParserFileInfo, err := subParserHub.DetermineFileTypeFromFile(subFile)
		if err != nil {
			log_helper.GetLogger().Errorln(err)
			continue
		}
		if subParserFileInfo == nil {
			// 说明这个字幕无法解析
			log_helper.GetLogger().Warnln("ReadSeriesInfoFromDir", seriesInfo.DirPath, "DetermineFileTypeFromFile is nil")
			continue
		}
		epsKey := pkg.GetEpisodeKeyName(info.Season, info.Episode)
		oneFileSubInfo := series.SubInfo{
			Title:        info.Title,
			Season:       info.Season,
			Episode:      info.Episode,
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
	// 视频字典 S01E01 - EpisodeInfo
	EpisodeDict := make(map[string]series.EpisodeInfo)
	for _, videoFile := range videoFiles {
		getEpsInfoAndSubDic(videoFile, EpisodeDict, SubDict)
	}

	for _, episodeInfo := range EpisodeDict {
		seriesInfo.EpList = append(seriesInfo.EpList, episodeInfo)
		seriesInfo.SeasonDict[episodeInfo.Season] = episodeInfo.Season
	}

	seriesInfo.NeedDlEpsKeyList, seriesInfo.NeedDlSeasonDict = whichSeasonEpsNeedDownloadSub(seriesInfo)

	return seriesInfo, nil
}

// ReadSeriesInfoFromEmby 将 Emby API 读取到的数据进行转换到通用的结构中，需要填充那些剧集需要下载，这样要的是一个连续剧的，不是所有的传入
func ReadSeriesInfoFromEmby(seriesDir string, imdbInfo *imdb.Title, seriesList []emby.EmbyMixInfo) (*series.SeriesInfo, error) {

	seriesInfo, err := getSeriesInfoFromDir(seriesDir, imdbInfo)
	if err != nil {
		return nil, err
	}
	seriesInfo.NeedDlSeasonDict = make(map[int]int)
	seriesInfo.NeedDlEpsKeyList = make(map[string]series.EpisodeInfo)
	EpisodeDict := make(map[string]series.EpisodeInfo)
	SubDict := make(map[string][]series.SubInfo)
	for _, info := range seriesList {
		getEpsInfoAndSubDic(info.VideoFileFullPath, EpisodeDict, SubDict)
	}

	for _, episodeInfo := range EpisodeDict {
		seriesInfo.EpList = append(seriesInfo.EpList, episodeInfo)
		seriesInfo.SeasonDict[episodeInfo.Season] = episodeInfo.Season
	}

	seriesInfo.NeedDlEpsKeyList, seriesInfo.NeedDlSeasonDict = whichSeasonEpsNeedDownloadSub(seriesInfo)

	return seriesInfo, nil
}

// SkipChineseSeries 跳过中文连续剧
func SkipChineseSeries(seriesRootPath string, _reqParam ...types.ReqParam) (bool, *imdb.Title, error) {
	var reqParam types.ReqParam
	if len(_reqParam) > 0 {
		reqParam = _reqParam[0]
	}
	imdbInfo, err := decode.GetImdbInfo4SeriesDir(seriesRootPath)
	if err != nil {
		return false, nil, err
	}
	t, err := imdb_helper.GetVideoInfoFromIMDB(imdbInfo.ImdbId, reqParam)
	if err != nil {
		return false, nil, err
	}
	if len(t.Languages) > 0 && strings.ToLower(t.Languages[0]) == "chinese" {
		log_helper.GetLogger().Infoln("Skip", filepath.Base(seriesRootPath), "Sub Download, because series is Chinese")
		return true, t, nil
	}
	return false, t, nil
}

// OneSeriesDlSubInAllSite 一部连续剧在所有的网站下载相应的字幕
func OneSeriesDlSubInAllSite(Suppliers []ifaces.ISupplier, seriesInfo *series.SeriesInfo, i int) []supplier.SubInfo {
	defer func() {
		log_helper.GetLogger().Infoln("DlSub End", seriesInfo.DirPath)
	}()
	log_helper.GetLogger().Infoln("DlSub Start", seriesInfo.DirPath)
	log_helper.GetLogger().Debugln(seriesInfo.Name, "IMDB ID:", seriesInfo.ImdbId, "NeedDownloadSubs:", len(seriesInfo.NeedDlEpsKeyList))
	var outSUbInfos = make([]supplier.SubInfo, 0)
	if len(seriesInfo.NeedDlEpsKeyList) < 1 {
		return outSUbInfos
	}
	for key, _ := range seriesInfo.NeedDlEpsKeyList {
		log_helper.GetLogger().Debugln(key)
	}
	// 同时进行查询
	subInfosChannel := make(chan []supplier.SubInfo)
	for _, oneSupplier := range Suppliers {
		nowSupplier := oneSupplier
		go func() {
			defer func() {
				log_helper.GetLogger().Infoln(i, nowSupplier.GetSupplierName(), "End...")
			}()
			log_helper.GetLogger().Infoln(i, nowSupplier.GetSupplierName(), "Start...")
			// 一次性把这一部连续剧的所有字幕下载完
			subInfos, err := nowSupplier.GetSubListFromFile4Series(seriesInfo)
			if err != nil {
				log_helper.GetLogger().Errorln(nowSupplier.GetSupplierName(), "GetSubListFromFile4Series", err)
			}
			// 把后缀名给改好
			sub_helper.ChangeVideoExt2SubExt(subInfos)

			subInfosChannel <- subInfos
		}()
	}
	for i := 0; i < len(Suppliers); i++ {
		v, ok := <-subInfosChannel
		if ok == true {
			outSUbInfos = append(outSUbInfos, v...)
		}
	}

	return outSUbInfos
}

// GetSeriesList 获取这个目录下的所有文件夹名称，默认为一个连续剧的目录的List
func GetSeriesList(dir string) ([]string, error) {

	var seriesDirList = make([]string, 0)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, curFile := range files {
		if curFile.IsDir() == false {

			// 如果发现有 tvshow.nfo 文件，那么就任务这个目录就是剧集的目录
			if strings.ToLower(curFile.Name()) == decode.MetadateTVNfo {
				seriesDirList = make([]string, 0)
				seriesDirList = append(seriesDirList, dir)
				return seriesDirList, nil
			}
			continue
		}
		fullPath := path.Join(dir, curFile.Name())
		seriesDirList = append(seriesDirList, fullPath)
	}

	return seriesDirList, err
}

// whichSeasonEpsNeedDownloadSub 有那些 Eps 需要下载的，按 SxEx 反回 epsKey
func whichSeasonEpsNeedDownloadSub(seriesInfo *series.SeriesInfo) (map[string]series.EpisodeInfo, map[int]int) {
	var needDlSubEpsList = make(map[string]series.EpisodeInfo, 0)
	var needDlSeasonList = make(map[int]int, 0)
	currentTime := time.Now()
	// 3个月
	dayRange, _ := time.ParseDuration(common.DownloadSubDuring3Months)
	for _, epsInfo := range seriesInfo.EpList {
		// 如果没有字幕，则加入下载列表
		// 如果每一集的播出时间能够读取到，那么就以这个完后推算 3个月
		// 如果读取不到 Aired Time 那么，这一集下载后的 ModifyTime 3个月天内，都进行字幕的下载
		var err error
		var baseTime time.Time
		if epsInfo.AiredTime != "" {
			baseTime, err = now.Parse(epsInfo.AiredTime)
			if err != nil {
				log_helper.GetLogger().Errorln("SeriesInfo parse AiredTime", err)
				baseTime = epsInfo.ModifyTime
			}
		} else {
			baseTime = epsInfo.ModifyTime
		}

		if len(epsInfo.SubAlreadyDownloadedList) < 1 || baseTime.Add(dayRange).After(currentTime) == true {
			// 添加
			epsKey := pkg.GetEpisodeKeyName(epsInfo.Season, epsInfo.Episode)
			needDlSubEpsList[epsKey] = epsInfo
			needDlSeasonList[epsInfo.Season] = epsInfo.Season
		} else {
			if len(epsInfo.SubAlreadyDownloadedList) > 0 {
				log_helper.GetLogger().Infoln("Skip because find sub file and downloaded or aired over 3 months,", epsInfo.Title, epsInfo.Season, epsInfo.Episode)
			} else if baseTime.Add(dayRange).After(currentTime) == false {
				log_helper.GetLogger().Infoln("Skip because 3 months pass,", epsInfo.Title, epsInfo.Season, epsInfo.Episode)
			}
		}
	}
	return needDlSubEpsList, needDlSeasonList
}

func getSeriesInfoFromDir(seriesDir string, imdbInfo *imdb.Title) (*series.SeriesInfo, error) {
	seriesInfo := series.SeriesInfo{}
	// 只考虑 IMDB 去查询，文件名目前发现可能会跟电影重复，导致很麻烦，本来也有前置要求要削刮器处理的
	videoInfo, err := decode.GetImdbInfo4SeriesDir(seriesDir)
	if err != nil {
		return nil, err
	}
	// 使用 IMDB ID 得到通用的剧集名称
	// 以 IMDB 的信息为准
	if imdbInfo != nil {
		seriesInfo.Name = imdbInfo.Name
		seriesInfo.ImdbId = imdbInfo.ID
		seriesInfo.Year = imdbInfo.Year
	} else {
		seriesInfo.Name = videoInfo.Title
		seriesInfo.ImdbId = videoInfo.ImdbId
		iYear, err := strconv.Atoi(videoInfo.Year)
		if err != nil {
			// 不是必须的
			seriesInfo.Year = 0
			log_helper.GetLogger().Warnln("ReadSeriesInfoFromDir.GetImdbInfo4SeriesDir.strconv.Atoi", err)
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

func getEpsInfoAndSubDic(videoFile string, EpisodeDict map[string]series.EpisodeInfo, SubDict map[string][]series.SubInfo) {
	// 正常来说，一集只有一个格式的视频，也就是 S01E01 只有一个，如果有多个则会只保存第一个
	info, modifyTime, err := decode.GetVideoInfoFromFileFullPath(videoFile)
	if err != nil {
		log_helper.GetLogger().Errorln("model.GetVideoInfoFromFileFullPath", err)
		return
	}
	episodeInfo, err := decode.GetImdbInfo4OneSeriesEpisode(videoFile)
	if err != nil {
		log_helper.GetLogger().Errorln("model.GetImdbInfo4OneSeriesEpisode", err)
		return
	}
	epsKey := pkg.GetEpisodeKeyName(info.Season, info.Episode)
	_, ok := EpisodeDict[epsKey]
	if ok == false {
		// 初始化
		oneFileEpInfo := series.EpisodeInfo{
			Title:        info.Title,
			Season:       info.Season,
			Episode:      info.Episode,
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
