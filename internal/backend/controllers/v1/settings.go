package v1

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/common"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/allanpk716/ChineseSubFinder/internal/types/backend"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (cb ControllerBase) SettingsHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "SettingsHandler", err)
	}()

	switch c.Request.Method {
	case "GET":
		{
			// 回复没有密码的 settings
			c.JSON(http.StatusOK, settings.GetSettings().GetNoPasswordSettings())
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
			err = settings.GetSettings().Read()
			if err != nil {
				return
			}
			nowPassword := settings.GetSettings().UserInfo.Password
			reqSetupInfo.UserInfo.Password = nowPassword
			err = settings.SetFullNewSettings(&reqSetupInfo)
			if err != nil {
				return
			}
			err = settings.GetSettings().Save()
			if err != nil {
				return
			}
			// ----------------------------------------
			// 设置接口的 API TOKEN
			if settings.GetSettings().ExperimentalFunction.ApiKeySettings.Enabled == true {
				common.SetApiToken(settings.GetSettings().ExperimentalFunction.ApiKeySettings.Key)
			} else {
				common.SetApiToken("")
			}
			// ----------------------------------------
			// 不管如何，都进行一次代理服务器的关闭，然后开启由具体的 获取 ProxySettings GetLocalHttpProxyUrl 操作开启这个服务器
			err = settings.GetSettings().AdvancedSettings.ProxySettings.CloseLocalHttpProxyServer()
			if err != nil {
				return
			}
			// 重新设置本地的静态文件服务器
			cb.StaticFileSystemBackEnd.Stop()
			cb.StaticFileSystemBackEnd.Start(cb.cronHelper.Settings.CommonSettings)

			c.JSON(http.StatusOK, backend.ReplyCommon{Message: "Settings Save Success"})
		}
	default:
		c.JSON(http.StatusNoContent, backend.ReplyCommon{Message: "Settings Request.Method Error"})
	}
}
