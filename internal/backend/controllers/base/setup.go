package base

import (
	"net/http"

	"github.com/allanpk716/ChineseSubFinder/internal/types/backend"
	"github.com/allanpk716/ChineseSubFinder/pkg/settings"
	"github.com/gin-gonic/gin"
)

func (cb ControllerBase) SetupHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "SetupHandler", err)
	}()

	setupInfo := backend.ReqSetupInfo{}
	err = c.ShouldBindJSON(&setupInfo)
	if err != nil {
		return
	}
	// 只有当用户不存在的时候才能够执行初始化操作
	found := false
	if settings.GetSettings().UserInfo.Username != "" && settings.GetSettings().UserInfo.Password != "" {
		found = true
	}

	if found == true {
		// 存在则反馈无需初始化
		c.JSON(http.StatusNoContent, backend.ReplyCommon{Message: "already setup"})
		return
	} else {
		// 需要创建用户，因为上述判断了没有用户存在，所以就默认直接新建了
		err = settings.SetFullNewSettings(&setupInfo.Settings)
		if err != nil {
			return
		}
		c.JSON(http.StatusOK, backend.ReplyCommon{Message: "ok"})
		return
	}
}
