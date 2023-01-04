package v1

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	backend2 "github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/backend"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"
	TTaskqueue "github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/task_queue"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/decode"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/video_scan_and_refresh_helper"
	"github.com/gin-gonic/gin"
)

func (cb *ControllerBase) RefreshVideoListStatusHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "RefreshVideoListStatusHandler", err)
	}()

	status := "running"
	if cb.videoScanAndRefreshHelperIsRunning == false {
		status = "stopped"
	}

	c.JSON(http.StatusOK, backend2.ReplyRefreshVideoList{
		Status:     status,
		ErrMessage: cb.videoScanAndRefreshHelperErrMessage})
	return
}

func (cb *ControllerBase) RefreshVideoListHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "RefreshVideoListHandler", err)
	}()

	if cb.videoScanAndRefreshHelperLocker.Lock() == false {
		// 已经在执行，跳过
		c.JSON(http.StatusOK, backend2.ReplyRefreshVideoList{
			Status: "running"})
		return
	}
	cb.videoScanAndRefreshHelper.NeedForcedScanAndDownSub = true
	cb.videoScanAndRefreshHelperIsRunning = true
	go func() {

		startT := time.Now()
		cb.log.Infoln("------------------------------------")
		cb.log.Infoln("Video Scan Started By webui...")

		pathUrlMap := cb.GetPathUrlMap()
		cb.log.Infoln("---------------------------------")
		cb.log.Infoln("GetPathUrlMap")
		for s, s2 := range pathUrlMap {
			cb.log.Infoln("pathUrlMap", s, s2)
		}
		cb.log.Infoln("---------------------------------")

		defer func() {
			cb.videoScanAndRefreshHelperIsRunning = false
			cb.videoScanAndRefreshHelperLocker.Unlock()
			cb.log.Infoln("Video Scan Finished By webui, cost:", time.Since(startT).Minutes(), "min")
			cb.log.Infoln("------------------------------------")
		}()
		// 先进行扫描
		var err2 error
		var scanVideoResult *video_scan_and_refresh_helper.ScanVideoResult
		cb.videoScanAndRefreshHelperErrMessage = ""
		scanVideoResult, err2 = cb.videoScanAndRefreshHelper.ScanNormalMovieAndSeries()
		if err2 != nil {
			cb.log.Errorln("ScanNormalMovieAndSeries", err2)
			cb.videoScanAndRefreshHelperErrMessage = err2.Error()
			return
		}
		err2 = cb.videoScanAndRefreshHelper.ScanEmbyMovieAndSeries(scanVideoResult)
		if err2 != nil {
			cb.log.Errorln("ScanEmbyMovieAndSeries", err2)
			cb.videoScanAndRefreshHelperErrMessage = err2.Error()
			return
		}

		MovieInfos, SeasonInfos := cb.videoScanAndRefreshHelper.ScrabbleUpVideoList(scanVideoResult, pathUrlMap)

		cb.log.Debugln("---------------------------------")
		for i, i2 := range MovieInfos {
			cb.log.Debugln("MovieInfos", i, i2.VideoFPath)
		}
		cb.log.Debugln("---------------------------------")
		for i, seasonInfo := range SeasonInfos {
			cb.log.Debugln("SeasonInfos.RootDirPath", i, seasonInfo.RootDirPath)
			for i2, i3 := range seasonInfo.OneVideoInfos {
				cb.log.Debugln("SeasonInfos.SeasonInfos", i2, i3.VideoFPath)
			}
		}
		cb.log.Debugln("---------------------------------")

		// 缓存视频列表
		cb.cronHelper.Downloader.SetMovieAndSeasonInfo(MovieInfos, SeasonInfos)
	}()

	c.JSON(http.StatusOK, backend2.ReplyRefreshVideoList{
		Status: "running"})
	return
}

func (cb *ControllerBase) VideoListAddHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "VideoListAddHandler", err)
	}()

	videoListAdd := backend2.ReqVideoListAdd{}
	err = c.ShouldBindJSON(&videoListAdd)
	if err != nil {
		return
	}

	videoType := common.Movie
	if videoListAdd.VideoType == 1 {
		videoType = common.Series
	}

	oneJob := TTaskqueue.NewOneJob(
		videoType, videoListAdd.PhysicalVideoFileFullPath, videoListAdd.TaskPriorityLevel,
		videoListAdd.MediaServerInsideVideoID,
	)

	if videoType == common.Series {
		// 如果是连续剧，需要额外的读取这一个剧集的信息
		epsVideoNfoInfo, err := decode.GetVideoNfoInfo4OneSeriesEpisode(videoListAdd.PhysicalVideoFileFullPath)
		if err != nil {
			return
		}
		seriesInfoDirPath := decode.GetSeriesDirRootFPath(videoListAdd.PhysicalVideoFileFullPath)
		if seriesInfoDirPath == "" {
			err = errors.New(fmt.Sprintf("decode.GetSeriesDirRootFPath == Empty, %s", videoListAdd.PhysicalVideoFileFullPath))
			return
		}
		oneJob.Season = epsVideoNfoInfo.Season
		oneJob.Episode = epsVideoNfoInfo.Episode
		oneJob.SeriesRootDirPath = seriesInfoDirPath
	}

	bok, err := cb.cronHelper.DownloadQueue.Add(*oneJob)
	if err != nil {
		return
	}
	if bok == false {
		// 任务已经存在
		bok, err = cb.cronHelper.DownloadQueue.Update(*oneJob)
		if err != nil {
			return
		}
		if bok == false {
			c.JSON(http.StatusOK, backend2.ReplyJobThings{
				JobID:   oneJob.Id,
				Message: "update job status failed",
			})
			return
		}
	}

	c.JSON(http.StatusOK, backend2.ReplyJobThings{
		JobID:   oneJob.Id,
		Message: "ok",
	})
}

func (cb *ControllerBase) VideoListHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "VideoListHandler", err)
	}()

	outMovieInfos, outSeasonInfo := cb.cronHelper.Downloader.GetMovieInfoAndSeasonInfo()

	c.JSON(http.StatusOK, backend2.ReplyVideoList{
		MovieInfos:  outMovieInfos,
		SeasonInfos: outSeasonInfo,
	})
}
