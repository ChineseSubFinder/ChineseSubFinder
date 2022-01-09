package router_v1

import (
	"github.com/allanpk716/ChineseSubFinder/internal/backend/middle"
	"github.com/allanpk716/ChineseSubFinder/internal/backend/routers/handler_v1"
	"github.com/gin-gonic/gin"
)

func GetRouters(router *gin.Engine) *gin.RouterGroup {

	router.Use(middle.CheckAuth())
	v1 := router.Group("/v1")
	{
		v1.POST("/login", handler_v1.PostLoginHandler)

		v1.POST("/change-pwd", handler_v1.PostChangePwdHandler)

		v1.GET("/settings", handler_v1.GetSettingsHandler)
		v1.PATCH("/settings", handler_v1.PatchSettingsHandler)

		v1.POST("/check-proxy", handler_v1.PostCheckProxyHandler)

		v1.POST("/check-path", handler_v1.PostCheckPathHandler)

		v1.POST("/jobs/start", handler_v1.PostJobStartHandler)
		v1.POST("/jobs/stop", handler_v1.PostJobStopHandler)
	}
	return v1
}
