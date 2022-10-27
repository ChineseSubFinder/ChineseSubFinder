package v1

import (
	"net/http"
	"strconv"

	"github.com/allanpk716/ChineseSubFinder/pkg"

	"github.com/allanpk716/ChineseSubFinder/pkg/preview_queue"
	backend2 "github.com/allanpk716/ChineseSubFinder/pkg/types/backend"
	"github.com/gin-gonic/gin"
)

// PreviewAdd 添加需要预览的任务
func (cb *ControllerBase) PreviewAdd(c *gin.Context) {

	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "PreviewAdd", err)
	}()

	job := preview_queue.Job{}
	err = c.ShouldBindJSON(&job)
	if err != nil {
		return
	}

	cb.cronHelper.Downloader.PreviewQueue.Add(&job)
	c.JSON(http.StatusOK, backend2.ReplyCommon{Message: "ok"})
	return
}

// PreviewList 列举预览任务
func (cb *ControllerBase) PreviewList(c *gin.Context) {

	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "PreviewList", err)
	}()

	listJob := cb.cronHelper.Downloader.PreviewQueue.ListJob()
	c.JSON(http.StatusOK, preview_queue.Reply{
		Jobs: listJob,
	})
}

// PreviewIsJobInQueue 预览的任务是否在列表中，或者说是在执行中
func (cb *ControllerBase) PreviewIsJobInQueue(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "PreviewIsJobInQueue", err)
	}()

	job := preview_queue.Job{}
	err = c.ShouldBindJSON(&job)
	if err != nil {
		return
	}

	found := cb.cronHelper.Downloader.PreviewQueue.IsJobInQueue(&preview_queue.Job{
		VideoFPath: job.VideoFPath,
	})

	c.JSON(http.StatusOK, backend2.ReplyCommon{Message: strconv.FormatBool(found)})
	return
}

// PreviewGetExportInfo 预览的任务的导出信息
func (cb *ControllerBase) PreviewGetExportInfo(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "PreviewGetExportInfo", err)
	}()

	job := preview_queue.Job{}
	err = c.ShouldBindJSON(&job)
	if err != nil {
		return
	}

	m3u8, subPath, err := cb.cronHelper.Downloader.PreviewQueue.GetVideoHLSAndSubByTimeRangeExportPathInfo(job.VideoFPath, job.SubFPath, job.StartTime, job.EndTime)
	if err != nil {
		return
	}

	c.JSON(http.StatusOK, preview_queue.Job{
		VideoFPath: m3u8,
		SubFPath:   subPath,
	})
	return
}

func (cb *ControllerBase) PreviewCleanUp(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "PreviewCleanUp", err)
	}()

	err = pkg.ClearVideoAndSubPreviewCacheFolder()
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, backend2.ReplyCommon{Message: "ok"})
}
