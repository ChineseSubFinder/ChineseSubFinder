package series_helper

import (
	"github.com/allanpk716/ChineseSubFinder/common"
	"github.com/allanpk716/ChineseSubFinder/model"
	"github.com/allanpk716/ChineseSubFinder/sub_parser/ass"
	"github.com/allanpk716/ChineseSubFinder/sub_parser/srt"
	"path/filepath"
	"strconv"
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

		info, err := model.GetVideoInfoFromFileName(subFile)
		if err != nil {
			model.GetLogger().Errorln(err)
			continue
		}
		subParserFileInfo, err := subParserHub.DetermineFileTypeFromFile(subFile)
		if err != nil {
			model.GetLogger().Errorln(err)
			continue
		}
		epsKey := "S" + strconv.Itoa(info.Season) + "E" +strconv.Itoa(info.Episode)
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
		info, err := model.GetVideoInfoFromFileName(videoFile)
		if err != nil {
			model.GetLogger().Errorln(err)
			continue
		}
		epsKey := "S" + strconv.Itoa(info.Season) + "E" +strconv.Itoa(info.Episode)
		_, ok := EpisodeDict[epsKey]
		if ok == false {
			// 初始化
			oneFileEpInfo := common.EpisodeInfo{
				Title: info.Title,
				Season: info.Season,
				Episode: info.Episode,
				Dir: filepath.Dir(videoFile),
				FileFullPath: videoFile,
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

	return &seriesInfo, nil
}
