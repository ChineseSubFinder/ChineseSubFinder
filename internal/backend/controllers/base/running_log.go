package base

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
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

	tmpOnceLogs := log_helper.GetRecentOnceLogs(reqRunningLog.TheLastFewTimes)

	replyOnceLog := backend.NewReplyRunningLog()
	replyOnceLog.RecentLogs = append(replyOnceLog.RecentLogs, tmpOnceLogs...)

	c.JSON(http.StatusOK, replyOnceLog)
}
