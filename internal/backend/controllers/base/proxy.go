package base

import (
	"github.com/allanpk716/ChineseSubFinder/internal/logic/file_downloader"
	subSupplier "github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier/shooter"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier/subhd"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier/xunlei"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier/zimuku"
	"github.com/allanpk716/ChineseSubFinder/internal/types/backend"
	"github.com/gin-gonic/gin"
	"github.com/huandu/go-clone"
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

	tmpFileDownloader := clone.Clone(cb.fileDownloader).(*file_downloader.FileDownloader)
	tmpFileDownloader.Settings.AdvancedSettings.ProxySettings = &checkProxy.ProxySettings
	tmpFileDownloader.Settings.AdvancedSettings.ProxySettings.UseProxy = true

	// 使用提交过来的这个代理地址，测试多个字幕网站的可用性
	subSupplierHub := subSupplier.NewSubSupplierHub(
		// 这里无需传递下载字幕的缓存实例
		zimuku.NewSupplier(tmpFileDownloader),
		xunlei.NewSupplier(tmpFileDownloader),
		shooter.NewSupplier(tmpFileDownloader),
		subhd.NewSupplier(tmpFileDownloader),
	)

	outStatus := subSupplierHub.CheckSubSiteStatus()

	c.JSON(http.StatusOK, outStatus)

}
