package v1

import "github.com/gin-gonic/gin"

func (cb ControllerBase) CheckPathHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "CheckPathHandler", err)
	}()
}
