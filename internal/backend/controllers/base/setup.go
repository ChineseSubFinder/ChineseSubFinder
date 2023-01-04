package base

import (
	"net/http"

	backend2 "github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/backend"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
	"github.com/gin-gonic/gin"
)

func (cb *ControllerBase) SetupHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "SetupHandler", err)
	}()

	setupInfo := backend2.ReqSetupInfo{}
	err = c.ShouldBindJSON(&setupInfo)
	if err != nil {
		return
	}
	// 只有当用户不存在的时候才能够执行初始化操作
	found := false
	if settings.Get().UserInfo.Username != "" && settings.Get().UserInfo.Password != "" {
		found = true
	}

	if found == true {
		// 存在则反馈无需初始化
		c.JSON(http.StatusNoContent, backend2.ReplyCommon{Message: "already setup"})
	} else {
		// 需要创建用户，因为上述判断了没有用户存在，所以就默认直接新建了
		err = settings.SetFullNewSettings(&setupInfo.Settings)
		if err != nil {
			return
		}
		c.JSON(http.StatusOK, backend2.ReplyCommon{Message: "ok"})
	}

	// 回复完毕后，发送重启 http server 的信号
	cb.restartSignal <- 1
}
