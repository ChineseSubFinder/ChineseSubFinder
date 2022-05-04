package v1

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/video_scan_and_refresh_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/types/backend"
	"github.com/allanpk716/ChineseSubFinder/internal/types/task_queue"
	"github.com/gin-gonic/gin"
	"net/http"
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

	c.JSON(http.StatusOK, backend.ReplyRefreshVideoList{
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
		c.JSON(http.StatusOK, backend.ReplyRefreshVideoList{
			Status: "running"})
		return
	}
	cb.videoScanAndRefreshHelper.NeedForcedScanAndDownSub = true
	cb.videoScanAndRefreshHelperIsRunning = true
	go func() {
		defer func() {
			cb.videoScanAndRefreshHelperIsRunning = false
			cb.videoScanAndRefreshHelperLocker.Unlock()
			cb.log.Infoln("Video Scan End By webui")
			cb.log.Infoln("------------------------------------")
		}()

		cb.log.Infoln("------------------------------------")
		cb.log.Infoln("Video Scan Started By webui...")
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

		pathUrlMap := cb.StaticFileSystemBackEnd.GetPathUrlMap()
		if cb.cronHelper.Settings.EmbySettings.Enable == true {
			// Emby 情况
			if scanVideoResult.Emby == nil {
				return
			}

		} else {
			// Normal 情况
			if scanVideoResult.Normal == nil {
				return
			}
			cb.MovieInfos, cb.SeasonInfos = cb.videoScanAndRefreshHelper.ScrabbleUpVideoList(scanVideoResult, pathUrlMap)
		}

		println("haha")
		// 这里会把得到的 Normal 和 Emby 的结果都放入 cb.scanVideoResult
		// 根据 用户的情况，选择行返回是 Emby Or Normal 的结果
		// 并且如果是 Emby 那么会在页面上出现一个刷新字幕列表的按钮（这个需要 Emby 中video 的 ID）
	}()

	c.JSON(http.StatusOK, backend.ReplyRefreshVideoList{
		Status: "running"})
	return
}

func (cb ControllerBase) MovieListHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "MovieListHandler", err)
	}()

	bok, allJobs, err := cb.cronHelper.DownloadQueue.GetAllJobs()
	if err != nil {
		return
	}

	if bok == false {
		c.JSON(http.StatusOK, backend.ReplyAllJobs{
			AllJobs: make([]task_queue.OneJob, 0),
		})
		return
	}

	c.JSON(http.StatusOK, backend.ReplyAllJobs{
		AllJobs: allJobs,
	})
}

func (cb ControllerBase) SeriesListHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "SeriesListHandler", err)
	}()

	bok, allJobs, err := cb.cronHelper.DownloadQueue.GetAllJobs()
	if err != nil {
		return
	}

	if bok == false {
		c.JSON(http.StatusOK, backend.ReplyAllJobs{
			AllJobs: make([]task_queue.OneJob, 0),
		})
		return
	}

	c.JSON(http.StatusOK, backend.ReplyAllJobs{
		AllJobs: allJobs,
	})
}
