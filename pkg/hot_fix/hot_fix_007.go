package hot_fix

import (
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"
	"github.com/sirupsen/logrus"
)

/*
	修复之前移除 subhd 和 zimuku 配置后，导致的无法获取字幕搜索链接的问题
*/
type HotFix007 struct {
	log *logrus.Logger
}

func NewHotFix007(log *logrus.Logger) *HotFix007 {
	return &HotFix007{log: log}
}

func (h HotFix007) GetKey() string {
	return "007"
}

func (h HotFix007) Process() (interface{}, error) {

	defer func() {
		h.log.Infoln("Hotfix", h.GetKey(), "End")
	}()

	h.log.Infoln("Hotfix", h.GetKey(), "Start...")

	return h.process()
}

func (h HotFix007) process() (bool, error) {

	if settings.Get().AdvancedSettings.SuppliersSettings.SubHD == nil {
		settings.Get().AdvancedSettings.SuppliersSettings.SubHD = settings.NewOneSupplierSettings(common.SubSiteSubHd, common.SubSubHDRootUrlDef, common.SubSubHDSearchUrl, 20)
	}

	if settings.Get().AdvancedSettings.SuppliersSettings.Zimuku == nil {
		settings.Get().AdvancedSettings.SuppliersSettings.Zimuku = settings.NewOneSupplierSettings(common.SubSiteZiMuKu, common.SubZiMuKuRootUrlDef, common.SubZiMuKuSearchFormatUrl, 20)
	}

	err := settings.Get().Save()
	if err != nil {
		return false, err
	}
	settings.Get(true)

	return true, nil
}
