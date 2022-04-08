package settings

import "github.com/allanpk716/ChineseSubFinder/internal/common"

type SuppliersSettings struct {
	Xunlei  *OneSupplierSettings `json:"xunlei"`
	Shooter *OneSupplierSettings `json:"shooter"`
	SubHD   *OneSupplierSettings `json:"subhd"`
	Zimuku  *OneSupplierSettings `json:"zimuku"`
}

func NewSuppliersSettings() *SuppliersSettings {
	return &SuppliersSettings{
		Xunlei:  NewOneSupplierSettings(common.SubSiteXunLei, common.SubXunLeiRootUrlDef, -1),
		Shooter: NewOneSupplierSettings(common.SubSiteShooter, common.SubShooterRootUrlDef, -1),
		SubHD:   NewOneSupplierSettings(common.SubSiteSubHd, common.SubSubHDRootUrlDef, 50),
		Zimuku:  NewOneSupplierSettings(common.SubSiteZiMuKu, common.SubZiMuKuRootUrlDef, 50),
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
