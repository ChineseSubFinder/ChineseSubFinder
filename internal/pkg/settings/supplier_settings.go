package settings

import (
	common2 "github.com/allanpk716/ChineseSubFinder/internal/types/common"
)

type SuppliersSettings struct {
	Xunlei  *OneSupplierSettings `json:"xunlei"`
	Shooter *OneSupplierSettings `json:"shooter"`
	SubHD   *OneSupplierSettings `json:"subhd"`
	Zimuku  *OneSupplierSettings `json:"zimuku"`
}

func NewSuppliersSettings() *SuppliersSettings {
	return &SuppliersSettings{
		Xunlei:  NewOneSupplierSettings(common2.SubSiteXunLei, common2.SubXunLeiRootUrlDef, -1),
		Shooter: NewOneSupplierSettings(common2.SubSiteShooter, common2.SubShooterRootUrlDef, -1),
		SubHD:   NewOneSupplierSettings(common2.SubSiteSubHd, common2.SubSubHDRootUrlDef, 50),
		Zimuku:  NewOneSupplierSettings(common2.SubSiteZiMuKu, common2.SubZiMuKuRootUrlDef, 50),
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
