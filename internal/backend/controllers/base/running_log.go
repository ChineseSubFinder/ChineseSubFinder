package base

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/types/backend"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func (cb ControllerBase) RunningLogHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "RunningLogHandler", err)
	}()

	// 从缓存中拿到日志信息，拼接后返回
	strLastFewTimes := c.DefaultQuery("the_last_few_times", "3")
	lastFewTimes, err := strconv.Atoi(strLastFewTimes)
	if err != nil {
		return
	}

	tmpOnceLogs := log_helper.GetRecentOnceLogs(lastFewTimes)

	replyOnceLog := backend.NewReplyRunningLog()
	replyOnceLog.RecentLogs = append(replyOnceLog.RecentLogs, tmpOnceLogs...)

	c.JSON(http.StatusOK, replyOnceLog)
}
