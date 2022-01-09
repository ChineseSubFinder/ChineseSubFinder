package router_v1

import "github.com/gin-gonic/gin"

func GetRouters(router *gin.Engine) *gin.RouterGroup {
	v1 := router.Group("/v1")
	//{
	//	v1.POST("/login", loginEndpoint)
	//	v1.POST("/submit", submitEndpoint)
	//	v1.POST("/read", readEndpoint)
	//}
	return v1
}
