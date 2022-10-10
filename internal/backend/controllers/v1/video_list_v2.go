package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/allanpk716/ChineseSubFinder/pkg/path_helper"
	backend2 "github.com/allanpk716/ChineseSubFinder/pkg/types/backend"
	vsh "github.com/allanpk716/ChineseSubFinder/pkg/video_scan_and_refresh_helper"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// RefreshMainList 重构后的视频列表，比如 x:\电影\壮志凌云\壮志凌云.mp4 或者是连续剧的 x:\连续剧\绝命毒师 根目录
func (cb *ControllerBase) RefreshMainList(c *gin.Context) {

	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "RefreshMainList", err)
	}()

	if cb.videoScanAndRefreshHelperLocker.Lock() == false {
		// 已经在执行，跳过
		c.JSON(http.StatusOK, backend2.ReplyRefreshVideoList{
			Status: "running"})
		return
	}
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

		var err2 error
		cb.videoScanAndRefreshHelperErrMessage = ""
		var mainList *vsh.NormalScanVideoResult
		mainList, err2 = cb.videoListHelper.RefreshMainList()
		if err2 != nil {
			cb.log.Errorln("RefreshMainList", err2)
			cb.videoScanAndRefreshHelperErrMessage = err2.Error()
			return
		}
		err2 = cb.cronHelper.Downloader.SetMovieAndSeasonInfoV2(mainList)
		if err2 != nil {
			cb.log.Errorln("SetMovieAndSeasonInfoV2", err2)
			cb.videoScanAndRefreshHelperErrMessage = err2.Error()
			return
		}
	}()

	c.JSON(http.StatusOK, backend2.ReplyRefreshVideoList{
		Status: "running"})
	return
}

// VideoMainList 获取电影和连续剧的基础结构
func (cb *ControllerBase) VideoMainList(c *gin.Context) {

	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "MoviePoster", err)
	}()
	outMovieInfos, outSeasonInfo, err := cb.cronHelper.Downloader.GetMovieInfoAndSeasonInfoV2()
	if err != nil {
		cb.log.Errorln("GetMovieInfoAndSeasonInfoV2", err)
		return
	}

	c.JSON(http.StatusOK, backend2.ReplyMainList{
		MovieInfos:  outMovieInfos,
		SeasonInfos: outSeasonInfo,
	})
}

// MoviePoster 获取电影海报
func (cb *ControllerBase) MoviePoster(c *gin.Context) {

	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "MoviePoster", err)
	}()

	movieInfo := backend2.MovieInfoV2{}
	err = c.ShouldBindJSON(&movieInfo)
	if err != nil {
		return
	}

	// 然后还需要将这个全路径信息转换为 静态文件服务器对应的路径返回给前端
	desUrl, found := cb.GetPathUrlMap()[movieInfo.MainRootDirFPath]
	if found == false {
		// 没有找到对应的 URL
		errMessage := fmt.Sprintf("MoviePoster.GetPathUrlMap can not find url for path %s", movieInfo.MainRootDirFPath)
		cb.log.Warningln(errMessage)
		err = errors.New(errMessage)
		return
	}
	posterFPath := cb.videoListHelper.GetMoviePoster(movieInfo.VideoFPath)
	posterUrl := path_helper.ChangePhysicalPathToSharePath(posterFPath, movieInfo.MainRootDirFPath, desUrl)

	c.JSON(http.StatusOK, backend2.PosterInfo{
		Url: posterUrl,
	})
}

// SeriesPoster 从一个连续剧的根目录中，获取连续剧的海报
func (cb *ControllerBase) SeriesPoster(c *gin.Context) {

	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "SeriesPoster", err)
	}()

	seriesInfo := backend2.SeasonInfoV2{}
	err = c.ShouldBindJSON(&seriesInfo)
	if err != nil {
		return
	}

	// 然后还需要将这个全路径信息转换为 静态文件服务器对应的路径返回给前端
	desUrl, found := cb.GetPathUrlMap()[seriesInfo.MainRootDirFPath]
	if found == false {
		// 没有找到对应的 URL
		errMessage := fmt.Sprintf("SeriesPoster.GetPathUrlMap can not find url for path %s", seriesInfo.MainRootDirFPath)
		cb.log.Warningln(errMessage)
		err = errors.New(errMessage)
		return
	}
	posterFPath := cb.videoListHelper.GetSeriesPoster(seriesInfo.RootDirPath)
	posterUrl := path_helper.ChangePhysicalPathToSharePath(posterFPath, seriesInfo.MainRootDirFPath, desUrl)

	c.JSON(http.StatusOK, backend2.PosterInfo{
		Url: posterUrl,
	})
}
