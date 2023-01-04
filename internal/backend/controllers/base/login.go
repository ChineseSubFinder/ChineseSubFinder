package base

import (
	"net/http"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	backend2 "github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/backend"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/common"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
	"github.com/gin-gonic/gin"
)

func (cb *ControllerBase) LoginHandler(c *gin.Context) {

	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "LoginHandler", err)
	}()
	nowUserInfo := settings.UserInfo{}
	err = c.ShouldBindJSON(&nowUserInfo)
	if err != nil {
		return
	}

	if settings.Get().UserInfo.Username == "" || settings.Get().UserInfo.Password == "" {
		// 配置文件中的账号和密码任意一个未空，提示用户需要进行 setup 流程
		c.JSON(http.StatusNoContent, backend2.ReplyCommon{Message: "You need do `Setup`"})
		return
	}

	if settings.Get().UserInfo.Username != nowUserInfo.Username ||
		settings.Get().UserInfo.Password != nowUserInfo.Password {
		// 账号密码不匹配
		c.JSON(http.StatusBadRequest, backend2.ReplyCommon{Message: "Username or Password Error"})
		return
	} else {
		// 用户账号密码匹配
		nowAccessToken := pkg.GenerateAccessToken()
		common.SetAccessToken(nowAccessToken)
		c.JSON(http.StatusOK, backend2.ReplyLogin{AccessToken: nowAccessToken,
			Settings: *settings.Get().GetNoPasswordSettings()})
		return
	}
}
