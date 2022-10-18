package v1

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	PTN "github.com/middelink/go-parse-torrent-name"

	"github.com/allanpk716/ChineseSubFinder/internal/models"

	"github.com/allanpk716/ChineseSubFinder/pkg/search"

	"github.com/allanpk716/ChineseSubFinder/pkg/sub_helper"

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

// OneMovieSubs 由一部电影去搜索其当前目录下的对应字幕
func (cb *ControllerBase) OneMovieSubs(c *gin.Context) {

	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "OneMovieSubs", err)
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
		errMessage := fmt.Sprintf("OneMovieSubs.GetPathUrlMap can not find url for path %s", movieInfo.MainRootDirFPath)
		cb.log.Warningln(errMessage)
		err = errors.New(errMessage)
		return
	}

	matchedSubs, err := sub_helper.SearchMatchedSubFileByOneVideo(cb.log, movieInfo.VideoFPath)
	if err != nil {
		cb.log.Errorln("OneMovieSubs.SearchMatchedSubFileByOneVideo", err)
		return
	}

	movieSubsInfo := backend2.MovieSubsInfo{
		SubUrlList: make([]string, 0),
	}
	// 将匹配到的字幕文件转换为 URL
	for _, sub := range matchedSubs {
		subUrl := path_helper.ChangePhysicalPathToSharePath(sub, movieInfo.MainRootDirFPath, desUrl)
		movieSubsInfo.SubUrlList = append(movieSubsInfo.SubUrlList, subUrl)
	}

	c.JSON(http.StatusOK, movieSubsInfo)
}

func (cb *ControllerBase) OneSeriesSubs(c *gin.Context) {

	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "OneSeriesSubs", err)
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
		errMessage := fmt.Sprintf("OneSeriesSubs.GetPathUrlMap can not find url for path %s", seriesInfo.MainRootDirFPath)
		cb.log.Warningln(errMessage)
		err = errors.New(errMessage)
		return
	}

	seasonInfo, err := search.SeriesAllEpsAndSubtitles(cb.log, seriesInfo.RootDirPath)
	if err != nil {
		cb.log.Errorln("OneSeriesSubs.SeriesAllEpsAndSubtitles", err)
		return
	}

	for i, videoInfo := range seasonInfo.OneVideoInfos {
		for _, subFPath := range videoInfo.SubFPathList {
			subUrl := path_helper.ChangePhysicalPathToSharePath(subFPath, seriesInfo.MainRootDirFPath, desUrl)
			seasonInfo.OneVideoInfos[i].SubUrlList = append(seasonInfo.OneVideoInfos[i].SubUrlList, subUrl)
		}
	}

	c.JSON(http.StatusOK, seasonInfo)
}

// ScanSkipInfo 设置或者获取跳过扫描信息的状态
func (cb *ControllerBase) ScanSkipInfo(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "ScanSkipInfo", err)
	}()

	switch c.Request.Method {
	case "POST":
		{
			// 查询
			videoSkipInfo := backend2.ReqVideoSkipInfo{}
			err = c.ShouldBindJSON(&videoSkipInfo)
			if err != nil {
				return
			}
			isSkip := cb.cronHelper.Downloader.ScanLogic.Get(videoSkipInfo.VideoType, videoSkipInfo.PhysicalVideoFileFullPath)
			c.JSON(http.StatusOK, backend2.ReplyVideoSkipInfo{
				IsSkip: isSkip,
			})
			return
		}
	case "PUT":
		{
			// 设置
			videoSkipInfo := backend2.ReqVideoSkipInfo{}
			err = c.ShouldBindJSON(&videoSkipInfo)
			if err != nil {
				return
			}

			var skipInfo *models.SkipScanInfo
			if videoSkipInfo.VideoType == 0 {
				// 电影
				skipInfo = models.NewSkipScanInfoByMovie(videoSkipInfo.PhysicalVideoFileFullPath, videoSkipInfo.IsSkip)
			} else {
				// 电视剧
				var parse *PTN.TorrentInfo
				parse, err = PTN.Parse(videoSkipInfo.PhysicalVideoFileFullPath)
				if err != nil {
					cb.log.Errorln("SetScanSkipInfo.PTN.Parse", err)
					return
				}
				dirFPath := filepath.Dir(filepath.Dir(videoSkipInfo.PhysicalVideoFileFullPath))
				skipInfo = models.NewSkipScanInfoBySeries(dirFPath, parse.Season, parse.Episode, videoSkipInfo.IsSkip)
			}

			cb.cronHelper.Downloader.ScanLogic.Set(skipInfo)

			c.JSON(http.StatusOK, backend2.ReplyCommon{
				Message: "ok"})
			return
		}
	default:
		c.JSON(http.StatusNoContent, backend2.ReplyCommon{Message: "ScanSkipInfo Request.Method Error"})
	}
}

// v.refreshEmbySubList() 可以使用这个方案去刷新最近视频的字幕列表，最近的多少条可以 V 来设置
