package v1

import (
	"fmt"
	"net/http"
	"path/filepath"

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

// ManualUploadSubtitle 人工上传字幕
func (cb *ControllerBase) ManualUploadSubtitle(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "ManualUploadSubtitle", err)
	}()
	videoFPath, ok := c.GetPostForm("video_f_path")
	if ok == false {
		err = fmt.Errorf("GetPostForm video_f_path failed")
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		cb.log.Errorln("ManualUploadSubtitle.FormFile", err)
		return
	}
	basePath := "./upload/"
	filename := basePath + filepath.Base(file.Filename)
	if err := c.SaveUploadedFile(file, filename); err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
		return
	}

	c.JSON(http.StatusOK, backend2.ReplyCommon{Message: "ok"})
	return
}
