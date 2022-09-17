package v1

import "github.com/gin-gonic/gin"

// ManualSubUploadOneHandler 一次上传一个字幕
func (cb *ControllerBase) ManualSubUploadOneHandler(c *gin.Context) {

	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "ManualSubUploadOneHandler", err)
	}()

}
