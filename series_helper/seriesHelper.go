package series_helper

import (
	"github.com/allanpk716/ChineseSubFinder/common"
	_interface "github.com/allanpk716/ChineseSubFinder/interface"
	"github.com/allanpk716/ChineseSubFinder/model"
	"github.com/allanpk716/ChineseSubFinder/sub_parser/ass"
	"github.com/allanpk716/ChineseSubFinder/sub_parser/srt"
	"path/filepath"
	"strings"
	"time"
)

// ReadSeriesInfoFromDir 读取剧集的信息
func ReadSeriesInfoFromDir(seriesDir string) (*common.SeriesInfo, error) {
	seriesInfo := common.SeriesInfo{}

	subParserHub := model.NewSubParserHub(ass.NewParser(), srt.NewParser())
	// 只考虑 IMDB 去查询，文件名目前发现可能会跟电影重复，导致很麻烦，本来也有前置要求要削刮器处理的
	videoInfo, err := model.GetImdbInfo(seriesDir)
	if err != nil {
		return nil, err
	}
	// 使用 IMDB ID 得到通用的剧集名称
	imdbInfo, err := model.GetVideoInfoFromIMDB(videoInfo.ImdbId)
	if err != nil {
		return nil, err
	}
	// 以 IMDB 的信息为准
	seriesInfo.Name = imdbInfo.Name
	seriesInfo.ImdbId = imdbInfo.ID
	seriesInfo.Year = imdbInfo.Year
	seriesInfo.DirPath = seriesDir
	seriesInfo.EpList = make([]common.EpisodeInfo, 0)
	seriesInfo.SeasonDict = make(map[int]int)
	// 搜索所有的视频
	videoFiles, err := model.SearchMatchedVideoFile(seriesDir)
	if err != nil {
		return nil, err
	}
	// 搜索所有的字幕
	subFiles, err := model.SearchMatchedSubFile(seriesDir)
	if err != nil {
		return nil, err
	}
	// 字幕字典 S01E01 - []SubInfo
	SubDict := make(map[string][]common.SubInfo)
	for _, subFile := range subFiles {

		info, _, err := model.GetVideoInfoFromFileFullPath(subFile)
		if err != nil {
			model.GetLogger().Errorln(err)
			continue
		}
		subParserFileInfo, err := subParserHub.DetermineFileTypeFromFile(subFile)
		if err != nil {
			model.GetLogger().Errorln(err)
			continue
		}
		epsKey := model.GetEpisodeKeyName(info.Season, info.Episode)
		oneFileSubInfo := common.SubInfo{
			Title: info.Title,
			Season: info.Season,
			Episode: info.Episode,
			Language: subParserFileInfo.Lang,
			Dir: filepath.Dir(subFile),
			FileFullPath: subFile,
		}
		_, ok := SubDict[epsKey]
		if ok == false {
			// 初始化
			SubDict[epsKey] = make([]common.SubInfo, 0)
		}
		SubDict[epsKey] = append(SubDict[epsKey], oneFileSubInfo)
	}
	// 视频字典 S01E01 - EpisodeInfo
	EpisodeDict := make(map[string]common.EpisodeInfo)
	for _, videoFile := range videoFiles {
		// 正常来说，一集只有一个格式的视频，也就是 S01E01 只有一个，如果有多个则会只保存第一个
		info, modifyTime, err := model.GetVideoInfoFromFileFullPath(videoFile)
		if err != nil {
			model.GetLogger().Errorln(err)
			continue
		}
		epsKey := model.GetEpisodeKeyName(info.Season, info.Episode)
		_, ok := EpisodeDict[epsKey]
		if ok == false {
			// 初始化
			oneFileEpInfo := common.EpisodeInfo{
				Title: info.Title,
				Season: info.Season,
				Episode: info.Episode,
				Dir: filepath.Dir(videoFile),
				FileFullPath: videoFile,
				ModifyTime: modifyTime,
			}
			// 需要匹配同级目录下的字幕
			oneFileEpInfo.SubList = make([]common.SubInfo, 0)
			for _, subInfo := range SubDict[epsKey] {
				if subInfo.Dir == oneFileEpInfo.Dir {
					oneFileEpInfo.SubList = append(oneFileEpInfo.SubList, subInfo)
				}
			}
			EpisodeDict[epsKey] = oneFileEpInfo
		} else {
			// 存在则跳过
			continue
		}
	}

	for _, episodeInfo := range EpisodeDict {
		seriesInfo.EpList = append(seriesInfo.EpList, episodeInfo)
		seriesInfo.SeasonDict[episodeInfo.Season] = episodeInfo.Season
	}

	seriesInfo.NeedDlEpsKeyList = whichEpsNeedDownloadSub(&seriesInfo)

	return &seriesInfo, nil
}

// SkipChineseSeries 跳过中文连续剧
func SkipChineseSeries(seriesRootPath string, _reqParam ...common.ReqParam) (bool, error) {
	var reqParam common.ReqParam
	if len(_reqParam) > 0 {
		reqParam = _reqParam[0]
	}
	imdbInfo, err := model.GetImdbInfo(seriesRootPath)
	if err != nil {
		return false, err
	}
	t, err := model.GetVideoInfoFromIMDB(imdbInfo.ImdbId, reqParam)
	if err != nil {
		return false, err
	}
	if len(t.Languages) > 0 && strings.ToLower(t.Languages[0]) == "chinese" {
		model.GetLogger().Infoln("Skip", filepath.Base(seriesRootPath), "Sub Download, because series is Chinese")
		return true, nil
	}
	return false, nil
}

// OneSeriesDlSubInAllSite 一部连续剧在所有的网站下载相应的字幕
func OneSeriesDlSubInAllSite(Suppliers []_interface.ISupplier, seriesInfo *common.SeriesInfo) []common.SupplierSubInfo {
	var outSUbInfos = make([]common.SupplierSubInfo, 0)
	// 同时进行查询
	subInfosChannel := make(chan []common.SupplierSubInfo)
	model.GetLogger().Infoln("DlSub Start", seriesInfo.DirPath)
	for _, supplier := range Suppliers {
		supplier := supplier
		go func() {
			subInfos, err := supplier.GetSubListFromFile4Series(seriesInfo)
			if err != nil {
				model.GetLogger().Errorln("GetSubListFromFile4Series", err)
			}
			subInfosChannel <- subInfos
		}()
	}
	for i := 0; i < len(Suppliers); i++ {
		v, ok := <-subInfosChannel
		if ok == true {
			outSUbInfos = append(outSUbInfos, v...)
		}
	}
	model.GetLogger().Infoln("DlSub End", seriesInfo.DirPath)
	return outSUbInfos
}

// whichEpsNeedDownloadSub 有那些 Eps 需要下载的，按 SxEx 反回 epsKey
func whichEpsNeedDownloadSub(seriesInfo *common.SeriesInfo) map[string]common.EpisodeInfo {
	var needDlSubEpsList = make(map[string]common.EpisodeInfo, 0)
	currentTime := time.Now()
	// 30 天
	dayRange, _ := time.ParseDuration(common.DownloadSubDuring30Days)
	for _, epsInfo := range seriesInfo.EpList {
		// 如果没有字幕，则加入下载列表
		// 这一集下载后的30天内，都进行字幕的下载
		if len(epsInfo.SubList) < 1 || epsInfo.ModifyTime.Add(dayRange).After(currentTime) == true {
			// 添加
			epsKey := model.GetEpisodeKeyName(epsInfo.Season, epsInfo.Episode)
			needDlSubEpsList[epsKey] = epsInfo
		} else {
			if len(epsInfo.SubList) > 0 {
				model.GetLogger().Infoln("Skip because find sub file", epsInfo.Title, epsInfo.Season, epsInfo.Episode)
			} else if epsInfo.ModifyTime.Add(dayRange).After(currentTime) == false {
				model.GetLogger().Infoln("Skip because 30 days pass", epsInfo.Title, epsInfo.Season, epsInfo.Episode)
			}
		}
	}
	return needDlSubEpsList
}