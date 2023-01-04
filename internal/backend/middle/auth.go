package middle

import (
	"net/http"
	"strings"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/backend"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/common"
	"github.com/gin-gonic/gin"
)

func CheckAuth() gin.HandlerFunc {

	return func(context *gin.Context) {
		authHeader := context.Request.Header.Get("Authorization")
		fields := strings.Fields(authHeader)
		if len(fields) != 2 {
			context.JSON(http.StatusUnauthorized, backend.ReplyCheckAuth{Message: "Request Header Authorization Error"})
			context.Abort()
			return
		}
		nowAccessToken := fields[1]
		if nowAccessToken == "" || nowAccessToken != common.GetAccessToken() {
			context.JSON(http.StatusUnauthorized, backend.ReplyCheckAuth{Message: "AccessToken Error"})
			context.Abort()
			return
		}
		// 向下传递消息
		context.Next()
	}
}

func CheckApiAuth() gin.HandlerFunc {

	return func(context *gin.Context) {
		authHeader := context.Request.Header.Get("Authorization")
		if len(authHeader) <= 1 {
			context.JSON(http.StatusUnauthorized, backend.ReplyCheckAuth{Message: "Request Header Authorization Error"})
			context.Abort()
			return
		}
		nowAccessToken := strings.Fields(authHeader)[1]
		if nowAccessToken == "" {
			context.JSON(http.StatusUnauthorized, backend.ReplyCheckAuth{Message: "api_key_enabled == false or api_key is empty"})
			context.Abort()
			return
		} else if nowAccessToken != common.GetApiToken() {
			context.JSON(http.StatusUnauthorized, backend.ReplyCheckAuth{Message: "AccessToken Error"})
			context.Abort()
			return
		}
		// 向下传递消息
		context.Next()
	}
}
