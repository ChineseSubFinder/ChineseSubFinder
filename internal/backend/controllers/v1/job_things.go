package v1

import (
	"github.com/allanpk716/ChineseSubFinder/internal/logic/cron_helper"
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
			cb.cronHelper.Start(settings.GetSettings().CommonSettings.RunScanAtStartUp)
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

	cronStatus := cb.cronHelper.CronHelperRunning()
	coreJobStatus := cb.cronHelper.FullDownloadProcessRunning()

	if coreJobStatus == true {
		// 核心任务在运行就是运行
		c.JSON(http.StatusOK, backend.ReplyJobStatus{
			Status: cron_helper.Running,
		})
	} else {
		// 核心任务没有运行，再判断是否定时器启动了
		if cronStatus == true {
			c.JSON(http.StatusOK, backend.ReplyJobStatus{
				Status: cron_helper.Running,
			})
		} else {
			c.JSON(http.StatusOK, backend.ReplyJobStatus{
				Status: cron_helper.Stopped,
			})
		}
	}

}
