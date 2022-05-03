package routers

import (
	"github.com/allanpk716/ChineseSubFinder/internal/backend/controllers/base"
	v1 "github.com/allanpk716/ChineseSubFinder/internal/backend/controllers/v1"
	"github.com/allanpk716/ChineseSubFinder/internal/backend/middle"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/cron_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/file_downloader"
	"github.com/gin-gonic/gin"
)

func InitRouter(fileDownloader *file_downloader.FileDownloader, router *gin.Engine, cronHelper *cron_helper.CronHelper) {

	cbBase := base.NewControllerBase(fileDownloader)
	cbV1 := v1.NewControllerBase(fileDownloader.Log, cronHelper)
	// 基础的路由
	router.GET("/system-status", cbBase.SystemStatusHandler)

	router.POST("/setup", cbBase.SetupHandler)

	router.POST("/login", cbBase.LoginHandler)
	router.POST("/logout", middle.CheckAuth(), cbBase.LogoutHandler)

	router.POST("/change-pwd", middle.CheckAuth(), cbBase.ChangePwdHandler)

	router.POST("/check-path", cbBase.CheckPathHandler)

	router.POST("/check-emby-path", cbBase.CheckEmbyPathHandler)

	router.POST("/check-proxy", cbBase.CheckProxyHandler)

	router.POST("/check-cron", cbBase.CheckCronHandler)

	router.GET("/def-settings", cbBase.DefSettingsHandler)

	// v1路由: /v1/xxx
	GroupV1 := router.Group("/" + cbV1.GetVersion())
	{
		GroupV1.Use(middle.CheckAuth())

		GroupV1.GET("/settings", cbV1.SettingsHandler)
		GroupV1.PUT("/settings", cbV1.SettingsHandler)

		GroupV1.POST("/daemon/start", cbV1.DaemonStartHandler)
		GroupV1.POST("/daemon/stop", cbV1.DaemonStopHandler)
		GroupV1.GET("/daemon/status", cbV1.DaemonStatusHandler)
	}
}
