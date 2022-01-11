package v1

import (
	"github.com/allanpk716/ChineseSubFinder/internal/dao"
	"github.com/allanpk716/ChineseSubFinder/internal/types/backend"
	"github.com/gin-gonic/gin"
	"net/http"
)

// SystemStatusHandler 获取系统状态
func (cb ControllerBase) SystemStatusHandler(c *gin.Context) {

	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "SystemStatusHandler", err)
	}()

	found, _, err := dao.GetUserInfo()
	if err != nil {
		return
	}

	c.JSON(http.StatusOK, backend.ReplySystemStatus{
		IsSetup: found,
	})
}
