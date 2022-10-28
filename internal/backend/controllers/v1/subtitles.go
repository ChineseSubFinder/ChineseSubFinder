package v1

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/allanpk716/ChineseSubFinder/pkg/manual_upload_sub_2_local"

	"github.com/allanpk716/ChineseSubFinder/pkg"

	backend2 "github.com/allanpk716/ChineseSubFinder/pkg/types/backend"
	"github.com/gin-gonic/gin"
)

// RefreshMediaServerSubList 刷新媒体服务器的字幕列表
func (cb *ControllerBase) RefreshMediaServerSubList(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "RefreshMediaServerSubList", err)
	}()

	err = cb.videoScanAndRefreshHelper.RefreshMediaServerSubList()
	if err != nil {
		cb.log.Errorln("RefreshMediaServerSubList", err)
		return
	}

	c.JSON(http.StatusOK, backend2.ReplyCommon{Message: "ok"})
	return
}

// ManualUploadSubtitle2Local 人工上传字幕到本地
func (cb *ControllerBase) ManualUploadSubtitle2Local(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "ManualUploadSubtitle2Local", err)
	}()
	videoFPath, ok := c.GetPostForm("video_f_path")
	if ok == false {
		err = fmt.Errorf("GetPostForm video_f_path failed")
		cb.log.Errorln(err)
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		err = fmt.Errorf("FormFile failed, err: %v", err)
		cb.log.Errorln(err)
		return
	}
	basePath, err := pkg.GetManualSubUploadCacheFolder()
	if err != nil {
		err = fmt.Errorf("GetManualSubUploadCacheFolder failed, err: %v", err)
		cb.log.Errorln(err)
		return
	}
	filename := filepath.Join(basePath, filepath.Base(file.Filename))
	if err = c.SaveUploadedFile(file, filename); err != nil {
		err = fmt.Errorf("SaveUploadedFile failed, err: %v", err)
		cb.log.Errorln(err)
		return
	}

	cb.cronHelper.Downloader.ManualUploadSub2Local.Add(&manual_upload_sub_2_local.Job{
		VideoFPath: videoFPath,
		SubFPath:   filename,
	})

	c.JSON(http.StatusOK, backend2.ReplyCommon{Message: "ok"})
	return
}

// ListManualUploadSubtitle2LocalJob 列举人工上传字幕到本地的任务列表
func (cb *ControllerBase) ListManualUploadSubtitle2LocalJob(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "ListManualUploadSubtitle2LocalJob", err)
	}()

	listJob := cb.cronHelper.Downloader.ManualUploadSub2Local.ListJob()
	c.JSON(http.StatusOK, manual_upload_sub_2_local.Reply{
		Jobs: listJob,
	})
	return
}

// IsManualUploadSubtitle2LocalJobInQueue 人工上传字幕到本地的任务是否在列表中，或者说是在执行中
func (cb *ControllerBase) IsManualUploadSubtitle2LocalJobInQueue(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "IsManualUploadSubtitle2LocalJobInQueue", err)
	}()

	job := manual_upload_sub_2_local.Job{}
	err = c.ShouldBindJSON(&job)
	if err != nil {
		return
	}

	found := cb.cronHelper.Downloader.ManualUploadSub2Local.IsJobInQueue(&manual_upload_sub_2_local.Job{
		VideoFPath: job.VideoFPath,
	})

	c.JSON(http.StatusOK, backend2.ReplyCommon{Message: strconv.FormatBool(found)})
	return
}

// ManualUploadSubtitleResult 人工上传字幕到本地的任务的结果，成功 ok，不存在空，其他是失败
func (cb *ControllerBase) ManualUploadSubtitleResult(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "ManualUploadSubtitleResult", err)
	}()

	job := manual_upload_sub_2_local.Job{}
	err = c.ShouldBindJSON(&job)
	if err != nil {
		return
	}

	result := cb.cronHelper.Downloader.ManualUploadSub2Local.JobResult(&manual_upload_sub_2_local.Job{
		VideoFPath: job.VideoFPath,
	})

	c.JSON(http.StatusOK, backend2.ReplyCommon{Message: result})
	return
}
