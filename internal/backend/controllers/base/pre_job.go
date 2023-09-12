package base

import (
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/backend"
	"github.com/gin-gonic/gin"
	"net/http"
)

// PreJobHandler 系统启动后的预处理工作
func (cb *ControllerBase) PreJobHandler(c *gin.Context) {

	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "PreJobHandler", err)
	}()

	c.JSON(http.StatusOK, backend.ReplyPreJob{
		IsDone:    cb.preJob.IsDone(),
		StageName: cb.preJob.GetStageName(),
	})
}
