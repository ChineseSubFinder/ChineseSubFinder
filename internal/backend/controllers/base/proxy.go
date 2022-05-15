package base

import (
	subSupplier "github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier/shooter"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier/subhd"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier/xunlei"
	"github.com/allanpk716/ChineseSubFinder/internal/types/backend"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (cb *ControllerBase) CheckProxyHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "CheckProxyHandler", err)
	}()

	checkProxy := backend.ReqCheckProxy{}
	err = c.ShouldBindJSON(&checkProxy)
	if err != nil {
		return
	}

	// 先尝试关闭之前的本地 http 代理
	err = cb.fileDownloader.Settings.AdvancedSettings.ProxySettings.CloseLocalHttpProxyServer()
	if err != nil {
		return
	}
	// 备份一份
	bkProxySettings := cb.fileDownloader.Settings.AdvancedSettings.ProxySettings.CopyOne()
	// 赋值 Web 传递过来的需要测试的代理参数
	cb.fileDownloader.Settings.AdvancedSettings.ProxySettings = &checkProxy.ProxySettings
	cb.fileDownloader.Settings.AdvancedSettings.ProxySettings.UseProxy = true

	// 使用提交过来的这个代理地址，测试多个字幕网站的可用性
	subSupplierHub := subSupplier.NewSubSupplierHub(
		// 这里无需传递下载字幕的缓存实例
		//zimuku.NewSupplier(cb.fileDownloader),
		xunlei.NewSupplier(cb.fileDownloader),
		shooter.NewSupplier(cb.fileDownloader),
		subhd.NewSupplier(cb.fileDownloader),
	)

	outStatus := subSupplierHub.CheckSubSiteStatus()

	defer func() {
		// 还原
		cb.fileDownloader.Settings.AdvancedSettings.ProxySettings = bkProxySettings
	}()

	c.JSON(http.StatusOK, outStatus)

}
