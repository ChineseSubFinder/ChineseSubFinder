package v1

import "github.com/gin-gonic/gin"

func (cb ControllerBase) JobStartHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "JobStartHandler", err)
	}()
}

func (cb ControllerBase) JobStopHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "JobStopHandler", err)
	}()

}
