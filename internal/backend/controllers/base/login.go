package base

import (
	"net/http"

	"github.com/allanpk716/ChineseSubFinder/internal/types/backend"
	"github.com/allanpk716/ChineseSubFinder/pkg/common"
	"github.com/allanpk716/ChineseSubFinder/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/pkg/settings"
	"github.com/gin-gonic/gin"
)

func (cb ControllerBase) LoginHandler(c *gin.Context) {

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

	if settings.GetSettings().UserInfo.Username == "" || settings.GetSettings().UserInfo.Password == "" {
		// 配置文件中的账号和密码任意一个未空，提示用户需要进行 setup 流程
		c.JSON(http.StatusNoContent, backend.ReplyCommon{Message: "You need do `Setup`"})
		return
	}

	if settings.GetSettings().UserInfo.Username != nowUserInfo.Username ||
		settings.GetSettings().UserInfo.Password != nowUserInfo.Password {
		// 账号密码不匹配
		c.JSON(http.StatusBadRequest, backend.ReplyCommon{Message: "Username or Password Error"})
		return
	} else {
		// 用户账号密码匹配
		nowAccessToken := my_util.GenerateAccessToken()
		common.SetAccessToken(nowAccessToken)
		c.JSON(http.StatusOK, backend.ReplyLogin{AccessToken: nowAccessToken,
			Settings: *settings.GetSettings().GetNoPasswordSettings()})
		return
	}
}
