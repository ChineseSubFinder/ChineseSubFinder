package backend

import (
	"fmt"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/pre_job"
	"net/http"

	"github.com/arl/statsviz"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/tmdb_api"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"

	"github.com/ChineseSubFinder/ChineseSubFinder/internal/backend/controllers/base"
	v1 "github.com/ChineseSubFinder/ChineseSubFinder/internal/backend/controllers/v1"
	"github.com/ChineseSubFinder/ChineseSubFinder/internal/backend/middle"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/cron_helper"
	"github.com/gin-gonic/gin"
)

func InitRouter(
	router *gin.Engine,
	cronHelper *cron_helper.CronHelper,
	restartSignal chan interface{},
	preJob *pre_job.PreJob,
) (*base.ControllerBase, *v1.ControllerBase) {

	// ----------------------------------------------
	// 设置 TMDB API 的本地 Client，用户自己的 API Key
	var err error
	var tmdbApi *tmdb_api.TmdbApi
	if settings.Get().AdvancedSettings.TmdbApiSettings.Enable == true &&
		settings.Get().AdvancedSettings.TmdbApiSettings.ApiKey != "" {

		tmdbApi, err = tmdb_api.NewTmdbHelper(cronHelper.Logger, settings.Get().AdvancedSettings.TmdbApiSettings.ApiKey, settings.Get().AdvancedSettings.TmdbApiSettings.UseAlternateBaseURL)
		if err != nil {
			cronHelper.Logger.Panicln("NewTmdbHelper", err)
		}
		if tmdbApi.Alive() == false {
			// 如果 tmdbApi 不可用，那么就不使用
			cronHelper.Logger.Errorln("tmdbApi.Alive() == false")
			tmdbApi = nil
		}
	}
	cronHelper.FileDownloader.MediaInfoDealers.SetTmdbHelperInstance(tmdbApi)
	// ----------------------------------------------
	cbBase := base.NewControllerBase(cronHelper.FileDownloader, restartSignal, preJob)
	cbV1 := v1.NewControllerBase(cronHelper, restartSignal)
	// --------------------------------------------------
	// 静态文件服务器
	// 添加电影的
	for i, path := range settings.Get().CommonSettings.MoviePaths {

		nowUrl := "/movie_dir_" + fmt.Sprintf("%d", i)
		cbV1.SetPathUrlMapItem(path, nowUrl)
		router.StaticFS(nowUrl, http.Dir(path))
	}
	// 添加连续剧的
	for i, path := range settings.Get().CommonSettings.SeriesPaths {

		nowUrl := "/series_dir_" + fmt.Sprintf("%d", i)
		cbV1.SetPathUrlMapItem(path, nowUrl)
		router.StaticFS(nowUrl, http.Dir(path))
	}
	// --------------------------------------------------
	// 性能监视
	if settings.Get().AdvancedSettings.DebugMode == true {
		// 如果是 DebugMode 那么开启性能监控
		router.GET("/debug/statsviz/*filepath", func(context *gin.Context) {
			if context.Param("filepath") == "/ws" {
				statsviz.Ws(context.Writer, context.Request)
				return
			}
			statsviz.IndexAtRoot("/debug/statsviz").ServeHTTP(context.Writer, context.Request)
		})
	}
	// --------------------------------------------------
	// 基础的路由
	router.GET("/system-status", cbBase.SystemStatusHandler)

	router.POST("/pre-job", cbBase.PreJobHandler)

	router.POST("/setup", cbBase.SetupHandler)

	router.POST("/login", cbBase.LoginHandler)
	router.POST("/logout", middle.CheckAuth(), cbBase.LogoutHandler)

	router.POST("/change-pwd", middle.CheckAuth(), cbBase.ChangePwdHandler)

	router.POST("/check-path", cbBase.CheckPathHandler)

	router.POST("/check-emby-path", cbBase.CheckEmbyPathHandler)

	router.POST("/check-proxy", cbBase.CheckProxyHandler)

	router.POST("/check-cron", cbBase.CheckCronHandler)

	router.GET("/def-settings", cbBase.DefSettingsHandler)

	router.POST("/check-emby-settings", cbBase.CheckEmbySettingsHandler)

	router.POST("/check-tmdb-api-settings", cbBase.CheckTmdbApiHandler)

	// v1路由: /v1/xxx
	GroupV1 := router.Group("/" + cbV1.GetVersion())
	{
		GroupV1.Use(middle.CheckAuth())

		GroupV1.GET("/settings", cbV1.SettingsHandler)
		GroupV1.PUT("/settings", cbV1.SettingsHandler)

		GroupV1.POST("/daemon/start", cbV1.DaemonStartHandler)
		GroupV1.POST("/daemon/stop", cbV1.DaemonStopHandler)
		GroupV1.GET("/daemon/status", cbV1.DaemonStatusHandler)

		GroupV1.GET("/jobs/list", cbV1.JobsListHandler)
		GroupV1.POST("/jobs/change-job-status", cbV1.ChangeJobStatusHandler)
		GroupV1.POST("/jobs/log", cbV1.JobLogHandler)

		//GroupV1.POST("/video/list/refresh", cbV1.RefreshVideoListHandler)
		GroupV1.GET("/video/list/refresh-status", cbV1.RefreshVideoListStatusHandler)
		//GroupV1.GET("/video/list", cbV1.VideoListHandler)
		GroupV1.POST("/video/list/add", cbV1.VideoListAddHandler)

		GroupV1.POST("/video/list/refresh_main_list", cbV1.RefreshMainList)
		GroupV1.GET("/video/list/video_main_list", cbV1.VideoMainList)
		GroupV1.POST("/video/list/movie_poster", cbV1.MoviePoster)
		GroupV1.POST("/video/list/series_poster", cbV1.SeriesPoster)
		GroupV1.POST("/video/list/one_movie_subs", cbV1.OneMovieSubs)
		GroupV1.POST("/video/list/one_series_subs", cbV1.OneSeriesSubs)
		GroupV1.POST("/video/list/scan_skip_info", cbV1.ScanSkipInfo)
		GroupV1.PUT("/video/list/scan_skip_info", cbV1.ScanSkipInfo)

		GroupV1.POST("/subtitles/refresh_media_server_sub_list", cbV1.RefreshMediaServerSubList)
		GroupV1.POST("/subtitles/manual_upload_2_local", cbV1.ManualUploadSubtitle2Local)
		GroupV1.POST("/subtitles/manual_upload_result", cbV1.ManualUploadSubtitleResult)
		GroupV1.GET("/subtitles/list_manual_upload_2_local_job", cbV1.ListManualUploadSubtitle2LocalJob)
		GroupV1.POST("/subtitles/is_manual_upload_2_local_in_queue", cbV1.IsManualUploadSubtitle2LocalJobInQueue)
		GroupV1.POST("/subtitles/get_generate_upload_url_info", cbV1.GetGenerateUploadURLHandle)

		GroupV1.POST("/preview/clean_up", cbV1.PreviewCleanUp)
		GroupV1.GET("/preview/playlist/:videofpathbase64", cbV1.HlsPlaylist)
		GroupV1.GET("/preview/segments/:resolution/:segment/:videofpathbase64", cbV1.HlsSegment)
		GroupV1.POST("/preview/search_other_web", cbV1.PreviewSearchOtherWeb)
		GroupV1.POST("/preview/video_f_path_2_imdb_info", cbV1.PreviewVideoFPath2IMDBInfo)
	}

	GroupAPIV1 := router.Group("/api/v1")
	{
		GroupAPIV1.Use(middle.CheckApiAuth())

		GroupAPIV1.POST("/add-job", cbV1.AddJobHandler)
		GroupAPIV1.GET("/job-status", cbV1.GetJobStatusHandler)
		GroupAPIV1.POST("/change-job-status", cbV1.ChangeJobStatusHandler)
		GroupAPIV1.POST("/add-video-played-info", cbV1.AddVideoPlayedInfoHandler)
		GroupAPIV1.DELETE("/del-video-played-info", cbV1.DelVideoPlayedInfoHandler)
	}

	return cbBase, cbV1
}
