package base

import (
	"net/http"
	"runtime"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/backend"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
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
	if settings.Get().UserInfo.Username != "" && settings.Get().UserInfo.Password != "" {
		// 进行过 setup 了，那么就可以 Login 的流程
		isSetup = true
	}

	c.JSON(http.StatusOK, backend.ReplySystemStatus{
		IsSetup:           isSetup,
		Version:           pkg.AppVersion(),
		OS:                runtime.GOOS,
		ARCH:              runtime.GOARCH,
		IsRunningInDocker: running.IsRunningInDocker()})
}
