package video_scan_and_refresh_helper

import (
	embyHelper "github.com/allanpk716/ChineseSubFinder/internal/logic/emby_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/forced_scan_and_down_sub"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/restore_fix_timeline_bk"
	seriesHelper "github.com/allanpk716/ChineseSubFinder/internal/logic/series_helper"
	subSupplier "github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier/zimuku"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/task_queue"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/decode"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	subTimelineFixerPKG "github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_timeline_fixer"
	"github.com/allanpk716/ChineseSubFinder/internal/types/common"
	"github.com/allanpk716/ChineseSubFinder/internal/types/emby"
	TTaskqueue "github.com/allanpk716/ChineseSubFinder/internal/types/task_queue"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/sirupsen/logrus"
)

type VideoScanAndRefreshHelper struct {
	settings                 *settings.Settings          // 设置的实例
	log                      *logrus.Logger              // 日志实例
	needForcedScanAndDownSub bool                        // 将会强制扫描所有的视频，下载字幕，替换已经存在的字幕，不进行时间段和已存在则跳过的判断。且不会进过 Emby API 的逻辑，智能进行强制去以本程序的方式去扫描。
	NeedRestoreFixTimeLineBK bool                        // 从 csf-bk 文件还原时间轴修复前的字幕文件
	embyHelper               *embyHelper.EmbyHelper      // Emby 的实例
	downloadQueue            *task_queue.TaskQueue       // 需要下载的视频的队列
	subSupplierHub           *subSupplier.SubSupplierHub // 字幕提供源的集合，仅仅是 check 是否需要下载字幕是足够的，如果要下载则需要额外的初始化和检查
}

func NewVideoScanAndRefreshHelper(settings *settings.Settings, log *logrus.Logger, downloadQueue *task_queue.TaskQueue) *VideoScanAndRefreshHelper {
	return &VideoScanAndRefreshHelper{settings: settings, log: log, downloadQueue: downloadQueue,
		subSupplierHub: subSupplier.NewSubSupplierHub(
			settings, log,
			zimuku.NewSupplier(settings, log),
		)}
}

// ReadSpeFile 优先级最高。读取特殊文件，启用一些特殊的功能，比如 forced_scan_and_down_sub
func (v *VideoScanAndRefreshHelper) ReadSpeFile() error {
	// 理论上是一次性的，用了这个文件就应该没了
	// 强制的字幕扫描
	needProcessForcedScanAndDownSub, err := forced_scan_and_down_sub.CheckSpeFile()
	if err != nil {
		return err
	}
	v.needForcedScanAndDownSub = needProcessForcedScanAndDownSub
	// 从 csf-bk 文件还原时间轴修复前的字幕文件
	needProcessRestoreFixTimelineBK, err := restore_fix_timeline_bk.CheckSpeFile()
	if err != nil {
		return err
	}
	v.NeedRestoreFixTimeLineBK = needProcessRestoreFixTimelineBK

	v.log.Infoln("NeedRestoreFixTimeLineBK ==", needProcessRestoreFixTimelineBK)

	return nil
}

// ScanMovieAndSeriesWait2DownloadSub 扫描出有那些电影、连续剧需要进行字幕下载的
func (v *VideoScanAndRefreshHelper) ScanMovieAndSeriesWait2DownloadSub() (*ScanVideoResult, error) {

	if v.settings.EmbySettings.Enable == false {
		v.embyHelper = nil

	} else {
		v.embyHelper = embyHelper.NewEmbyHelper(v.settings.EmbySettings)
	}

	var err error
	// -----------------------------------------------------
	// 强制下载和常规模式（没有媒体服务器）
	if v.needForcedScanAndDownSub == true || v.embyHelper == nil {

		normalScanResult := NormalScanVideoResult{}
		// 直接由本程序自己去扫描视频视频有哪些
		// 全扫描
		if v.needForcedScanAndDownSub == true {
			v.log.Infoln("Forced Scan And DownSub")
		}
		// --------------------------------------------------
		// 电影
		// 没有填写 emby_helper api 的信息，那么就走常规的全文件扫描流程
		normalScanResult.MovieFileFullPathList, err = my_util.SearchMatchedVideoFileFromDirs(v.log, v.settings.CommonSettings.MoviePaths)
		if err != nil {
			return nil, err
		}
		// --------------------------------------------------
		// 连续剧
		// 遍历连续剧总目录下的第一层目录
		normalScanResult.SeriesDirMap, err = seriesHelper.GetSeriesListFromDirs(v.settings.CommonSettings.SeriesPaths)
		if err != nil {
			return nil, err
		}
		// ------------------------------------------------------------------------------
		// 输出调试信息，有那些连续剧文件夹名称
		normalScanResult.SeriesDirMap.Each(func(key interface{}, value interface{}) {
			for i, s := range value.([]string) {
				v.log.Debugln("embyHelper == nil GetSeriesList", i, s)
			}
		})
		// ------------------------------------------------------------------------------
		return &ScanVideoResult{Normal: &normalScanResult}, nil
	} else {
		// TODO 如果后续支持了 Jellyfin、Plex 那么这里需要额外正在对应的扫描逻辑
		// 进过 emby_helper api 的信息读取
		embyScanResult := EmbyScanVideoResult{}
		v.log.Infoln("Movie Sub Dl From Emby API...")
		// Emby 情况，从 Emby 获取视频信息
		err = v.RefreshEmbySubList()
		if err != nil {
			v.log.Errorln("RefreshEmbySubList", err)
			return nil, err
		}
		// ------------------------------------------------------------------------------
		// 有哪些更新的视频列表，包含电影、连续剧
		embyScanResult.MovieSubNeedDlEmbyMixInfoList, embyScanResult.SeriesSubNeedDlEmbyMixInfoMap, err = v.GetUpdateVideoListFromEmby()
		if err != nil {
			v.log.Errorln("GetUpdateVideoListFromEmby", err)
			return nil, err
		}
		// ------------------------------------------------------------------------------
		return &ScanVideoResult{Emby: &embyScanResult}, nil
	}
}

// FilterMovieAndSeriesNeedDownload 过滤出需要下载字幕的视频，比如是否跳过中文的剧集，是否超过3个月的下载时间，丢入队列中
func (v *VideoScanAndRefreshHelper) FilterMovieAndSeriesNeedDownload(scanVideoResult *ScanVideoResult) error {

	if scanVideoResult.Normal != nil {
		err := v.filterMovieAndSeriesNeedDownloadNormal(scanVideoResult.Normal)
		if err != nil {
			return err
		}
	}

	if scanVideoResult.Emby != nil {
		err := v.filterMovieAndSeriesNeedDownloadEmby(scanVideoResult.Emby)
		if err != nil {
			return err
		}
	}

	return nil
}

func (v *VideoScanAndRefreshHelper) filterMovieAndSeriesNeedDownloadNormal(normal *NormalScanVideoResult) error {
	// ----------------------------------------
	// Normal 过滤，电影
	for _, oneMovieFPath := range normal.MovieFileFullPathList {
		// 放入队列
		if v.subSupplierHub.MovieNeedDlSub(oneMovieFPath, v.needForcedScanAndDownSub) == false {
			continue
		}

		bok, err := v.downloadQueue.Add(*TTaskqueue.NewOneJob(
			common.Movie, oneMovieFPath, task_queue.DefaultTaskPriorityLevel,
		))
		if err != nil {
			v.log.Errorln("filterMovieAndSeriesNeedDownloadNormal.Movie.NewOneJob", err)
			continue
		}
		if bok == false {
			v.log.Warningln("filterMovieAndSeriesNeedDownloadNormal", common.Movie.String(), oneMovieFPath, "downloadQueue.Add == false")
		}
	}
	// Normal 过滤，连续剧
	// seriesDirMap: dir <--> seriesList
	normal.SeriesDirMap.Each(func(seriesRootPathName interface{}, seriesNames interface{}) {

		for _, oneSeriesRootDir := range seriesNames.([]string) {

			// 因为可能回去 Web 获取 IMDB 信息，所以这里的错误不返回
			bNeedDlSub, seriesInfo, err := v.subSupplierHub.SeriesNeedDlSub(oneSeriesRootDir, v.needForcedScanAndDownSub)
			if err != nil {
				v.log.Errorln("filterMovieAndSeriesNeedDownloadNormal.SeriesNeedDlSub", err)
				continue
			}
			if bNeedDlSub == false {
				continue
			}

			for _, episodeInfo := range seriesInfo.NeedDlEpsKeyList {
				// 放入队列
				oneJob := TTaskqueue.NewOneJob(
					common.Series, episodeInfo.FileFullPath, task_queue.DefaultTaskPriorityLevel,
				)
				oneJob.Season = episodeInfo.Season
				oneJob.Episode = episodeInfo.Episode
				oneJob.SeriesRootDirPath = seriesInfo.DirPath

				bok, err := v.downloadQueue.Add(*oneJob)
				if err != nil {
					v.log.Errorln("filterMovieAndSeriesNeedDownloadNormal.Series.NewOneJob", err)
					continue
				}
				if bok == false {
					v.log.Warningln("filterMovieAndSeriesNeedDownloadNormal", common.Series.String(), episodeInfo.FileFullPath, "downloadQueue.Add == false")
				}
			}
		}
	})

	return nil
}

func (v *VideoScanAndRefreshHelper) filterMovieAndSeriesNeedDownloadEmby(emby *EmbyScanVideoResult) error {
	// ----------------------------------------
	// Emby 过滤，电影
	for _, oneMovieMixInfo := range emby.MovieSubNeedDlEmbyMixInfoList {
		// 放入队列
		if v.subSupplierHub.MovieNeedDlSub(oneMovieMixInfo.PhysicalVideoFileFullPath, v.needForcedScanAndDownSub) == false {
			continue
		}
		bok, err := v.downloadQueue.Add(*TTaskqueue.NewOneJob(
			common.Movie, oneMovieMixInfo.PhysicalVideoFileFullPath, task_queue.DefaultTaskPriorityLevel,
			oneMovieMixInfo.VideoInfo.Id,
		))
		if err != nil {
			v.log.Errorln("filterMovieAndSeriesNeedDownloadEmby.Movie.NewOneJob", err)
			continue
		}
		if bok == false {
			v.log.Warningln("filterMovieAndSeriesNeedDownloadEmby", common.Movie.String(), oneMovieMixInfo.PhysicalVideoFileFullPath, "downloadQueue.Add == false")
		}
	}
	// Emby 过滤，连续剧
	for _, embyMixInfos := range emby.SeriesSubNeedDlEmbyMixInfoMap {

		if len(embyMixInfos) < 1 {
			continue
		}

		// 只需要从一集取信息即可
		for _, mixInfo := range embyMixInfos {

			// 放入队列
			oneJob := TTaskqueue.NewOneJob(
				common.Series, mixInfo.PhysicalVideoFileFullPath, task_queue.DefaultTaskPriorityLevel,
				mixInfo.VideoInfo.Id,
			)

			info, _, err := decode.GetVideoInfoFromFileFullPath(mixInfo.PhysicalVideoFileFullPath)
			if err != nil {
				v.log.Errorln("filterMovieAndSeriesNeedDownloadEmby.Series.GetVideoInfoFromFileFullPath", err)
				continue
			}
			oneJob.Season = info.Season
			oneJob.Episode = info.Episode
			oneJob.SeriesRootDirPath = mixInfo.PhysicalSeriesRootDir

			bok, err := v.downloadQueue.Add(*oneJob)
			if err != nil {
				v.log.Errorln("filterMovieAndSeriesNeedDownloadEmby.Series.NewOneJob", err)
				continue
			}
			if bok == false {
				v.log.Warningln("filterMovieAndSeriesNeedDownloadEmby", common.Series.String(), mixInfo.PhysicalVideoFileFullPath, "downloadQueue.Add == false")
			}
		}
	}

	return nil
}

// GetUpdateVideoListFromEmby 这里首先会进行近期影片的获取，然后对这些影片进行刷新，然后在获取字幕列表，最终得到需要字幕获取的 video 列表
func (v *VideoScanAndRefreshHelper) GetUpdateVideoListFromEmby() ([]emby.EmbyMixInfo, map[string][]emby.EmbyMixInfo, error) {
	if v.embyHelper == nil {
		return nil, nil, nil
	}
	defer func() {
		v.log.Infoln("GetUpdateVideoListFromEmby End")
	}()
	v.log.Infoln("GetUpdateVideoListFromEmby Start...")
	//------------------------------------------------------
	var err error
	var movieList []emby.EmbyMixInfo
	var seriesSubNeedDlMap map[string][]emby.EmbyMixInfo //  多个需要搜索字幕的连续剧目录，连续剧文件夹名称 -- 每一集的 EmbyMixInfo List
	movieList, seriesSubNeedDlMap, err = v.embyHelper.GetRecentlyAddVideoListWithNoChineseSubtitle()
	if err != nil {
		return nil, nil, err
	}
	// 输出调试信息
	v.log.Debugln("GetUpdateVideoListFromEmby - DebugInfo - movieFileFullPathList Start")
	for _, info := range movieList {
		v.log.Debugln(info.PhysicalVideoFileFullPath)
	}
	v.log.Debugln("GetUpdateVideoListFromEmby - DebugInfo - movieFileFullPathList End")

	v.log.Debugln("GetUpdateVideoListFromEmby - DebugInfo - seriesSubNeedDlMap Start")
	for s := range seriesSubNeedDlMap {
		v.log.Debugln(s)
	}
	v.log.Debugln("GetUpdateVideoListFromEmby - DebugInfo - seriesSubNeedDlMap End")

	return movieList, seriesSubNeedDlMap, nil
}

func (v *VideoScanAndRefreshHelper) RefreshEmbySubList() error {

	if v.embyHelper == nil {
		return nil
	}

	bRefresh := false
	defer func() {
		if bRefresh == true {
			v.log.Infoln("Refresh Emby Sub List Success")
		} else {
			v.log.Errorln("Refresh Emby Sub List Error")
		}
	}()
	v.log.Infoln("Refresh Emby Sub List Start...")
	//------------------------------------------------------
	bRefresh, err := v.embyHelper.RefreshEmbySubList()
	if err != nil {
		return err
	}

	return nil
}

func (v *VideoScanAndRefreshHelper) RestoreFixTimelineBK() error {

	defer v.log.Infoln("End Restore Fix Timeline BK")
	v.log.Infoln("Start Restore Fix Timeline BK...")
	//------------------------------------------------------
	_, err := subTimelineFixerPKG.Restore(v.settings.CommonSettings.MoviePaths, v.settings.CommonSettings.SeriesPaths)
	if err != nil {
		return err
	}
	return nil
}

type ScanVideoResult struct {
	Normal *NormalScanVideoResult
	Emby   *EmbyScanVideoResult
}

type NormalScanVideoResult struct {
	MovieFileFullPathList []string
	SeriesDirMap          *treemap.Map
}

type EmbyScanVideoResult struct {
	MovieSubNeedDlEmbyMixInfoList []emby.EmbyMixInfo
	SeriesSubNeedDlEmbyMixInfoMap map[string][]emby.EmbyMixInfo
}
