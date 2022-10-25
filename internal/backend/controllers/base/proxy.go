package base

import (
	"net/http"

	"github.com/allanpk716/ChineseSubFinder/pkg/local_http_proxy_server"

	"github.com/allanpk716/ChineseSubFinder/pkg/settings"

	"github.com/allanpk716/ChineseSubFinder/pkg/logic/sub_supplier/a4k"

	"github.com/allanpk716/ChineseSubFinder/pkg/types/backend"

	subSupplier "github.com/allanpk716/ChineseSubFinder/pkg/logic/sub_supplier"
	"github.com/allanpk716/ChineseSubFinder/pkg/logic/sub_supplier/assrt"
	"github.com/allanpk716/ChineseSubFinder/pkg/logic/sub_supplier/csf"
	"github.com/allanpk716/ChineseSubFinder/pkg/logic/sub_supplier/shooter"
	"github.com/allanpk716/ChineseSubFinder/pkg/logic/sub_supplier/xunlei"

	"github.com/gin-gonic/gin"
)

func (cb *ControllerBase) CheckProxyHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "CheckProxyHandler", err)
	}()

	if cb.proxyCheckLocker.Lock() == false {
		c.JSON(http.StatusOK, backend.ReplyCommon{Message: "running"})
		return
	}

	defer func() {
		cb.proxyCheckLocker.Unlock()
	}()

	checkProxy := backend.ReqCheckProxy{}
	err = c.ShouldBindJSON(&checkProxy)
	if err != nil {
		return
	}
	// 备份一份
	bkProxySettings := settings.Get().AdvancedSettings.ProxySettings.CopyOne()
	// 赋值 Web 传递过来的需要测试的代理参数
	settings.Get().AdvancedSettings.ProxySettings = &checkProxy.ProxySettings
	settings.Get().AdvancedSettings.ProxySettings.UseProxy = true

	defer func() {
		// 还原
		settings.Get().AdvancedSettings.ProxySettings = bkProxySettings
		err = local_http_proxy_server.SetProxyInfo(settings.Get().AdvancedSettings.ProxySettings.GetInfos())
		if err != nil {
			return
		}
		// 启动代理
		local_http_proxy_server.GetProxyUrl()
	}()

	// 设置代理
	err = local_http_proxy_server.SetProxyInfo(settings.Get().AdvancedSettings.ProxySettings.GetInfos())
	if err != nil {
		return
	}
	// 使用提交过来的这个代理地址，测试多个字幕网站的可用性
	subSupplierHub := subSupplier.NewSubSupplierHub(
		// 这里无需传递下载字幕的缓存实例
		//zimuku.NewSupplier(cb.fileDownloader),
		//subhd.NewSupplier(cb.fileDownloader),
		xunlei.NewSupplier(cb.fileDownloader),
		shooter.NewSupplier(cb.fileDownloader),
		a4k.NewSupplier(cb.fileDownloader),
	)

	if settings.Get().SubtitleSources.AssrtSettings.Enabled == true &&
		settings.Get().SubtitleSources.AssrtSettings.Token != "" {
		// 如果开启了 ASSRt 字幕源，则需要测试 ASSRt 的代理
		subSupplierHub.AddSubSupplier(assrt.NewSupplier(cb.fileDownloader))
	}

	if settings.Get().ExperimentalFunction.ShareSubSettings.ShareSubEnabled == true {
		// 如果开启了分享字幕功能，那么就可以开启这个功能
		subSupplierHub.AddSubSupplier(csf.NewSupplier(cb.fileDownloader))
	}

	outStatus := subSupplierHub.CheckSubSiteStatus()

	c.JSON(http.StatusOK, outStatus)
}
