package routers

import (
	"github.com/allanpk716/ChineseSubFinder/internal/backend/controllers/base"
	v1 "github.com/allanpk716/ChineseSubFinder/internal/backend/controllers/v1"
	"github.com/allanpk716/ChineseSubFinder/internal/backend/middle"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/cron_helper"
	"github.com/gin-gonic/gin"
)

func InitRouter(router *gin.Engine, cronHelper *cron_helper.CronHelper) {

	cbBase := base.NewControllerBase()
	cbV1 := v1.NewControllerBase(cronHelper)
	// 基础的路由
	router.GET("/system-status", cbBase.SystemStatusHandler)

	router.POST("/setup", cbBase.SetupHandler)

	router.POST("/login", cbBase.LoginHandler)
	router.POST("/logout", middle.CheckAuth(), cbBase.LogoutHandler)

	router.POST("/change-pwd", middle.CheckAuth(), cbBase.ChangePwdHandler)

	router.POST("/check-path", cbBase.CheckPathHandler)

	router.POST("/check-proxy", cbBase.CheckProxyHandler)

	router.GET("/def-settings", cbBase.DefSettingsHandler)

	router.GET("/running-log", middle.CheckAuth(), cbBase.RunningLogHandler)

	// v1路由: /v1/xxx
	GroupV1 := router.Group("/" + cbV1.GetVersion())
	{
		GroupV1.Use(middle.CheckAuth())

		GroupV1.GET("/settings", cbV1.SettingsHandler)
		GroupV1.PUT("/settings", cbV1.SettingsHandler)

		GroupV1.POST("/jobs/start", cbV1.JobStartHandler)
		GroupV1.POST("/jobs/stop", cbV1.JobStopHandler)
		GroupV1.GET("/jobs/status", cbV1.JobStatusHandler)
	}
}
