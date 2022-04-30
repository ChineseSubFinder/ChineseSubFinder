package base

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/global_value"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
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

	isSetup := false
	if settings.GetSettings().UserInfo.Username != "" && settings.GetSettings().UserInfo.Password != "" {
		// 进行过 setup 了，那么就可以 Login 的流程
		isSetup = true
	}

	c.JSON(http.StatusOK, backend.ReplySystemStatus{
		IsSetup: isSetup,
		Version: global_value.AppVersion(),
	})
}
