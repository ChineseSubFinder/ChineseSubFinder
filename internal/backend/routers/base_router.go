package routers

import (
	v1 "github.com/allanpk716/ChineseSubFinder/internal/backend/controllers/v1"
	"github.com/allanpk716/ChineseSubFinder/internal/backend/middle"
	"github.com/gin-gonic/gin"
)

func InitRouter(router *gin.Engine) {

	cbV1 := v1.NewControllerBase()

	router.GET("/system-status", cbV1.SystemStatusHandler)

	router.POST("/setup", cbV1.SetupHandler)

	router.POST("/login", cbV1.LoginHandler)

	GroupV1 := router.Group("/" + cbV1.GetVersion())
	{
		GroupV1.Use(middle.CheckAuth())

		GroupV1.POST("/logout", cbV1.LogoutHandler)

		GroupV1.POST("/change-pwd", cbV1.ChangePwdHandler)

		GroupV1.GET("/settings", cbV1.SettingsHandler)
		GroupV1.PATCH("/settings", cbV1.SettingsHandler)

		GroupV1.POST("/check-proxy", cbV1.CheckProxyHandler)

		GroupV1.POST("/check-path", cbV1.CheckPathHandler)
		GroupV1.POST("/jobs/start", cbV1.JobStartHandler)
		GroupV1.POST("/jobs/stop", cbV1.JobStopHandler)
	}
}
