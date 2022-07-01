package settings

import (
	"github.com/allanpk716/ChineseSubFinder/pkg/types/common"
)

type SuppliersSettings struct {
	ChineseSubFinder *OneSupplierSettings `json:"csf"`
	Xunlei           *OneSupplierSettings `json:"xunlei"`
	Shooter          *OneSupplierSettings `json:"shooter"`
	SubHD            *OneSupplierSettings `json:"subhd"`
	Zimuku           *OneSupplierSettings `json:"zimuku"`
	Assrt            *OneSupplierSettings `json:"assrt"`
	A4k              *OneSupplierSettings `json:"a4k"`
}

func NewSuppliersSettings() *SuppliersSettings {
	return &SuppliersSettings{
		ChineseSubFinder: NewOneSupplierSettings(common.SubSiteChineseSubFinder, "haha", -1),
		Xunlei:           NewOneSupplierSettings(common.SubSiteXunLei, common.SubXunLeiRootUrlDef, -1),
		Shooter:          NewOneSupplierSettings(common.SubSiteShooter, common.SubShooterRootUrlDef, -1),
		SubHD:            NewOneSupplierSettings(common.SubSiteSubHd, common.SubSubHDRootUrlDef, 20),
		Zimuku:           NewOneSupplierSettings(common.SubSiteZiMuKu, common.SubZiMuKuRootUrlDef, 20),
		Assrt:            NewOneSupplierSettings(common.SubSiteAssrt, common.SubAssrtRootUrlDef, -1),
		A4k:              NewOneSupplierSettings(common.SubSiteA4K, common.SubA4kRootUrlDef, -1),
	}
}

type OneSupplierSettings struct {
	Name               string `json:"name"`
	RootUrl            string `json:"root_url"`
	DailyDownloadLimit int    `json:"daily_download_limit" default:"-1"` // -1 是无限制
}

func NewOneSupplierSettings(name string, rootUrl string, dailyDownloadLimit int) *OneSupplierSettings {
	return &OneSupplierSettings{Name: name, RootUrl: rootUrl, DailyDownloadLimit: dailyDownloadLimit}
}
