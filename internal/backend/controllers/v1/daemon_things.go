package v1

import (
	"net/http"

	backend2 "github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/backend"

	"github.com/gin-gonic/gin"
)

func (cb *ControllerBase) DaemonStartHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "DaemonStartHandler", err)
	}()

	if cb.cronHelper.CronHelperRunning() == false {
		go func() {
			// 砍掉，启动就进行扫描的逻辑
			cb.cronHelper.Start(false)
		}()
	}

	c.JSON(http.StatusOK, backend2.ReplyCommon{
		Message: "ok",
	})
}

func (cb *ControllerBase) DaemonStopHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "DaemonStopHandler", err)
	}()

	if cb.cronHelper.CronHelperRunning() == true {
		go func() {
			cb.cronHelper.Stop()
		}()
	}

	c.JSON(http.StatusOK, backend2.ReplyCommon{
		Message: "ok",
	})
}

func (cb *ControllerBase) DaemonStatusHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "DaemonStatusHandler", err)
	}()

	c.JSON(http.StatusOK, backend2.ReplyJobStatus{
		Status: cb.cronHelper.CronRunningStatusString(),
	})
}
