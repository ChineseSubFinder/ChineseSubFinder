package v1

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/allanpk716/ChineseSubFinder/internal/types/backend"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (cb ControllerBase) JobStartHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "JobStartHandler", err)
	}()

	if cb.cronHelper.CronHelperRunning() == false {
		go func() {
			cb.cronHelper.Start(settings.GetSettings(true).CommonSettings.RunScanAtStartUp)
		}()
	}

	c.JSON(http.StatusOK, backend.ReplyCommon{
		Message: "ok",
	})
}

func (cb ControllerBase) JobStopHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "JobStopHandler", err)
	}()

	if cb.cronHelper.CronHelperRunning() == true {
		go func() {
			cb.cronHelper.Stop()
		}()
	}

	c.JSON(http.StatusOK, backend.ReplyCommon{
		Message: "ok",
	})
}

func (cb ControllerBase) JobStatusHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "JobStatusHandler", err)
	}()

	c.JSON(http.StatusOK, backend.ReplyJobStatus{
		Status: cb.cronHelper.CronRunningStatusString(),
	})
}
