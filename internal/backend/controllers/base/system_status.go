package base

import (
	"github.com/allanpk716/ChineseSubFinder/pkg/types/backend"
	"net/http"
	"runtime"

	"github.com/allanpk716/ChineseSubFinder/pkg/global_value"
	"github.com/allanpk716/ChineseSubFinder/pkg/settings"
	running "github.com/allanpk716/is_running_in_docker"
	"github.com/gin-gonic/gin"
)

// SystemStatusHandler 获取系统状态
func (cb *ControllerBase) SystemStatusHandler(c *gin.Context) {

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
		IsSetup:           isSetup,
		Version:           global_value.AppVersion(),
		OS:                runtime.GOOS,
		ARCH:              runtime.GOARCH,
		IsRunningInDocker: running.IsRunningInDocker()})
}
