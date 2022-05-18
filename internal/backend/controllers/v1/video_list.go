package v1

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/allanpk716/ChineseSubFinder/internal/pkg/decode"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/video_scan_and_refresh_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/types/backend"
	"github.com/allanpk716/ChineseSubFinder/internal/types/common"
	TTaskqueue "github.com/allanpk716/ChineseSubFinder/internal/types/task_queue"
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

		MovieInfos, SeasonInfos := cb.videoScanAndRefreshHelper.ScrabbleUpVideoList(scanVideoResult, pathUrlMap)

		// 缓存视频列表
		cb.cronHelper.Downloader.SetMovieAndSeasonInfo(MovieInfos, SeasonInfos)
	}()

	c.JSON(http.StatusOK, backend.ReplyRefreshVideoList{
		Status: "running"})
	return
}

func (cb *ControllerBase) VideoListAddHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "VideoListAddHandler", err)
	}()

	videoListAdd := backend.ReqVideoListAdd{}
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

	torrentInfo, err := decode.GetVideoInfoFromFileName(videoListAdd.PhysicalVideoFileFullPath)
	if err != nil {
		return
	}

	seriesInfoDirPath := decode.GetSeriesDirRootFPath(videoListAdd.PhysicalVideoFileFullPath)
	if seriesInfoDirPath == "" {
		err = errors.New(fmt.Sprintf("decode.GetSeriesDirRootFPath == Empty, %s", videoListAdd.PhysicalVideoFileFullPath))
		return
	}

	oneJob.Season = torrentInfo.Season
	oneJob.Episode = torrentInfo.Episode
	oneJob.SeriesRootDirPath = seriesInfoDirPath

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
			c.JSON(http.StatusOK, backend.ReplyJobThings{
				JobID:   oneJob.Id,
				Message: "update job status failed",
			})
			return
		}
	}

	c.JSON(http.StatusOK, backend.ReplyJobThings{
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

	c.JSON(http.StatusOK, backend.ReplyVideoList{
		MovieInfos:  outMovieInfos,
		SeasonInfos: outSeasonInfo,
	})
}
