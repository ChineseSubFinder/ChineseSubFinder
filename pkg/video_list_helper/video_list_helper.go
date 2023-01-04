package video_list_helper

import (
	"sync"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/search"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
	vsh "github.com/ChineseSubFinder/ChineseSubFinder/pkg/video_scan_and_refresh_helper"

	seriesHelper "github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/series_helper"
	"github.com/sirupsen/logrus"
)

type VideoListHelper struct {
	log *logrus.Logger // 日志实例
}

func NewVideoListHelper(log *logrus.Logger) *VideoListHelper {
	return &VideoListHelper{
		log: log,
	}
}

// RefreshMainList 获取到电影、连续剧一级目录信息，不包含 Season 及其以下信息
// 只给出 Movie 的FullPath，Series 的 RootDirPath
func (v *VideoListHelper) RefreshMainList() (*vsh.NormalScanVideoResult, error) {

	defer func() {
		v.log.Infoln("ScanNormalMovieAndSeries End")
	}()
	v.log.Infoln("ScanNormalMovieAndSeries Start...")

	// ------------------------------------------------------------------------------
	// 由于需要进行视频信息的缓存，用于后续的逻辑，那么本地视频的扫描默认都会进行
	normalScanResult := vsh.NormalScanVideoResult{}
	// 直接由本程序自己去扫描视频视频有哪些
	// 全扫描
	wg := sync.WaitGroup{}
	var errMovie, errSeries error
	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
		}()
		// --------------------------------------------------
		// 电影
		normalScanResult.MoviesDirMap, errMovie = search.MatchedVideoFileFromDirs(v.log, settings.Get().CommonSettings.MoviePaths)
	}()
	wg.Add(1)
	go func() {

		defer func() {
			wg.Done()
		}()
		// --------------------------------------------------
		// 连续剧
		// 遍历连续剧总目录下的第一层目录
		normalScanResult.SeriesDirMap, errSeries = seriesHelper.GetSeriesListFromDirs(v.log, settings.Get().CommonSettings.SeriesPaths)
		// ------------------------------------------------------------------------------
		// 输出调试信息，有那些连续剧文件夹名称
		if normalScanResult.SeriesDirMap == nil {
			return
		}
		normalScanResult.SeriesDirMap.Each(func(key interface{}, value interface{}) {
			for i, s := range value.([]string) {
				v.log.Debugln("embyHelper == nil GetSeriesList", i, s)
			}
		})
	}()
	wg.Wait()
	if errMovie != nil {
		return nil, errMovie
	}
	if errSeries != nil {
		return nil, errSeries
	}

	return &normalScanResult, nil
}
