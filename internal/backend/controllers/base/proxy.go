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
	"github.com/huandu/go-clone"
	"net/http"
)

func (cb ControllerBase) CheckProxyHandler(c *gin.Context) {
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

	tmpSettings := clone.Clone(*settings.GetSettings()).(settings.Settings)
	tmpSettings.AdvancedSettings.ProxySettings.HttpProxyAddress = checkProxy.HttpProxyAddress

	// 使用提交过来的这个代理地址，测试多个字幕网站的可用性
	subSupplierHub := subSupplier.NewSubSupplierHub(
		tmpSettings,
		zimuku.NewSupplier(tmpSettings),
		xunlei.NewSupplier(tmpSettings),
		shooter.NewSupplier(tmpSettings),
		subhd.NewSupplier(tmpSettings),
	)

	outStatus := subSupplierHub.CheckSubSiteStatus()

	c.JSON(http.StatusOK, outStatus)

}
