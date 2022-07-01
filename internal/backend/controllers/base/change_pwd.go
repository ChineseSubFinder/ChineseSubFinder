package base

import (
	"net/http"

	"github.com/allanpk716/ChineseSubFinder/internal/types/backend"
	"github.com/allanpk716/ChineseSubFinder/pkg/common"
	"github.com/allanpk716/ChineseSubFinder/pkg/settings"
	"github.com/gin-gonic/gin"
)

func (cb ControllerBase) ChangePwdHandler(c *gin.Context) {

	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "ChangePwdHandler", err)
	}()

	changePwd := backend.ReqChangePwd{}
	err = c.ShouldBindJSON(&changePwd)
	if err != nil {
		return
	}

	if settings.GetSettings().UserInfo.Username == "" || settings.GetSettings().UserInfo.Password == "" {
		// 配置文件中的账号和密码任意一个未空，提示用户需要进行 setup 流程
		c.JSON(http.StatusNoContent, backend.ReplyCommon{Message: "You need do `Setup`"})
		return
	}

	if settings.GetSettings().UserInfo.Password != changePwd.OrgPwd {
		// 原始的密码不对
		c.JSON(http.StatusNoContent, backend.ReplyCommon{Message: "Org Password Error"})
	} else {
		// 同意修改密码
		settings.GetSettings().UserInfo.Password = changePwd.NewPwd
		err = settings.GetSettings().Save()
		if err != nil {
			return
		}
		// 修改密码成功后，会清理 AccessToken，强制要求重写登录
		common.SetAccessToken("")
		c.JSON(http.StatusOK, backend.ReplyCommon{Message: "ok, need ReLogin"})
	}
}
