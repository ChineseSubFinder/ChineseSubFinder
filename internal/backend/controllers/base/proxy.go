package base

import (
	subSupplier "github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier/shooter"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier/subhd"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier/xunlei"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier/zimuku"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
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

	tmpSettings := settings.GetSettings()
	tmpSettings.AdvancedSettings.ProxySettings.UseHttpProxy = true
	tmpSettings.AdvancedSettings.ProxySettings.HttpProxyAddress = checkProxy.HttpProxyAddress

	// 使用提交过来的这个代理地址，测试多个字幕网站的可用性
	subSupplierHub := subSupplier.NewSubSupplierHub(
		tmpSettings,
		cb.log,
		// 这里无需传递下载字幕的缓存实例
		zimuku.NewSupplier(tmpSettings, cb.log, nil),
		xunlei.NewSupplier(tmpSettings, cb.log, nil),
		shooter.NewSupplier(tmpSettings, cb.log, nil),
		subhd.NewSupplier(tmpSettings, cb.log, nil),
	)

	outStatus := subSupplierHub.CheckSubSiteStatus()

	c.JSON(http.StatusOK, outStatus)

}
