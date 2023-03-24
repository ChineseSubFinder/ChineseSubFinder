package base

import (
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_supplier/subtitle_best"
	"net/http"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/local_http_proxy_server"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_supplier/a4k"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/backend"

	subSupplier "github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_supplier"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_supplier/assrt"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_supplier/shooter"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_supplier/xunlei"

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

	if settings.Get().SubtitleSources.SubtitleBestSettings.Enabled == true &&
		settings.Get().SubtitleSources.SubtitleBestSettings.ApiKey != "" {
		// 如果开启了 SubtitleBest 字幕源，则需要测试 ASSRt 的代理
		subSupplierHub.AddSubSupplier(subtitle_best.NewSupplier(cb.fileDownloader))
	}

	outStatus := subSupplierHub.CheckSubSiteStatus()

	c.JSON(http.StatusOK, outStatus)
}
