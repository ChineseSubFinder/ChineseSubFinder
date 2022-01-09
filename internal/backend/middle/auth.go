package middle

import (
	"github.com/allanpk716/ChineseSubFinder/internal/types/backend"
	"github.com/gin-gonic/gin"
	"net/http"
)

func CheckAuth() gin.HandlerFunc {

	return func(context *gin.Context) {
		authHeader := context.Request.Header.Get("AccessToken")
		if authHeader != "123456" {
			context.JSON(http.StatusUnauthorized, backend.ReplyCheckAuth{Message: "Need Login!"})
			context.Abort()
			return
		}
		// 向下传递消息
		context.Next()
	}
}
