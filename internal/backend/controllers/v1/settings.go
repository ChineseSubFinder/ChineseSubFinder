package v1

import (
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"
	"net/http"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/backend"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/common"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
	"github.com/gin-gonic/gin"
)

func (cb *ControllerBase) SettingsHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "SettingsHandler", err)
	}()

	switch c.Request.Method {
	case "GET":
		{
			// 回复没有密码的 settings
			c.JSON(http.StatusOK, settings.Get().GetNoPasswordSettings())
		}
	case "PUT":
		{
			// 修改设置，这里不允许修改密码
			reqSetupInfo := settings.Settings{}
			err = c.ShouldBindJSON(&reqSetupInfo)
			if err != nil {
				return
			}
			// 需要去除 user 的 password 信息再保存，也就是继承之前的 password 即可
			nowPassword := settings.Get().UserInfo.Password
			reqSetupInfo.UserInfo.Password = nowPassword
			err = settings.SetFullNewSettings(&reqSetupInfo)
			if err != nil {
				return
			}
			pkg.ResetWantedVideoExt()
			// ----------------------------------------
			// 设置接口的 API TOKEN
			if settings.Get().ExperimentalFunction.ApiKeySettings.Enabled == true {
				common.SetApiToken(settings.Get().ExperimentalFunction.ApiKeySettings.Key)
			} else {
				common.SetApiToken("")
			}
			// ----------------------------------------
			c.JSON(http.StatusOK, backend.ReplyCommon{Message: "Settings Save Success"})
			// 回复完毕后，发送重启 http server 的信号
			cb.restartSignal <- 1
		}
	default:
		c.JSON(http.StatusNoContent, backend.ReplyCommon{Message: "Settings Request.Method Error"})
	}
}
