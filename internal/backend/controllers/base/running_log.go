package base

import (
	"github.com/allanpk716/ChineseSubFinder/internal/types/backend"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (cb ControllerBase) RunningLogHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "RunningLogHandler", err)
	}()

	reqRunningLog := backend.ReqRunningLog{}
	err = c.ShouldBindJSON(&reqRunningLog)
	if err != nil {
		return
	}
	// 从缓存中拿到日志信息，拼接后返回

	c.JSON(http.StatusOK, backend.NewReplyRunningLog())
}
