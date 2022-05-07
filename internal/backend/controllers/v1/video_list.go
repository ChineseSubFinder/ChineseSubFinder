package v1

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/video_scan_and_refresh_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/types/backend"
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

		cb.MovieInfos = make([]backend.MovieInfo, 0)
		cb.SeasonInfos = make([]backend.SeasonInfo, 0)
		MovieInfos, SeasonInfos := cb.videoScanAndRefreshHelper.ScrabbleUpVideoList(scanVideoResult, pathUrlMap)

		cb.MovieInfos = append(cb.MovieInfos, MovieInfos...)
		cb.SeasonInfos = append(cb.SeasonInfos, SeasonInfos...)
		// 并且如果是 Emby 那么会在页面上出现一个刷新字幕列表的按钮（这个需要 Emby 中video 的 ID）
	}()

	c.JSON(http.StatusOK, backend.ReplyRefreshVideoList{
		Status: "running"})
	return
}

func (cb *ControllerBase) VideoListHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "VideoListHandler", err)
	}()

	c.JSON(http.StatusOK, backend.ReplyVideoList{
		MovieInfos:  cb.MovieInfos,
		SeasonInfos: cb.SeasonInfos,
	})
}
