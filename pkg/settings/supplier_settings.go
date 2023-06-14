package settings

import (
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"
)

type SuppliersSettings struct {
	Xunlei       *OneSupplierSettings `json:"xunlei"`
	Shooter      *OneSupplierSettings `json:"shooter"`
	Assrt        *OneSupplierSettings `json:"assrt"`
	A4k          *OneSupplierSettings `json:"a4k"`
	SubHD        *OneSupplierSettings `json:"subhd"`
	Zimuku       *OneSupplierSettings `json:"zimuku"`
	SubtitleBest *OneSupplierSettings `json:"subtitle_best"`
}

func NewSuppliersSettings() *SuppliersSettings {
	return &SuppliersSettings{
		Xunlei:       NewOneSupplierSettings(common.SubSiteXunLei, common.SubXunLeiRootUrlDef, "", -1),
		Shooter:      NewOneSupplierSettings(common.SubSiteShooter, common.SubShooterRootUrlDef, "", -1),
		Assrt:        NewOneSupplierSettings(common.SubSiteAssrt, common.SubAssrtRootUrlDef, "", -1),
		A4k:          NewOneSupplierSettings(common.SubSiteA4K, common.SubA4kRootUrlDef, common.SubA4kSearchUrl, -1),
		SubtitleBest: NewOneSupplierSettings(common.SubSiteSubtitleBest, common.SubSubtitleBestRootUrlDef, common.SubSubtitleBestSearchMovieUrl, -1),
		// 依然需要给出来，用于手动搜索字幕使用
		SubHD:  NewOneSupplierSettings(common.SubSiteSubHd, common.SubSubHDRootUrlDef, common.SubSubHDSearchUrl, 20),
		Zimuku: NewOneSupplierSettings(common.SubSiteZiMuKu, common.SubZiMuKuRootUrlDef, common.SubZiMuKuSearchFormatUrl, 20),
	}
}

// ReSetSearchUrl 因为 SuppliersSettings 中每个网站的 searchUrl 参数没有开放更改，所以如果有变动，需要重新设置
func (s *SuppliersSettings) ReSetSearchUrl() {
	s.A4k.SearchUrl = common.SubA4kSearchUrl
	s.SubtitleBest.SearchUrl = common.SubSubtitleBestSearchMovieUrl
	s.SubHD.SearchUrl = common.SubSubHDSearchUrl
	s.Zimuku.SearchUrl = common.SubZiMuKuSearchFormatUrl
}

type OneSupplierSettings struct {
	Name               string `json:"name"`
	RootUrl            string `json:"root_url"`
	SearchUrl          string `json:"search_url"`
	DailyDownloadLimit int    `json:"daily_download_limit" default:"-1"` // -1 是无限制
}

func NewOneSupplierSettings(name string, rootUrl, searchUrl string, dailyDownloadLimit int) *OneSupplierSettings {
	return &OneSupplierSettings{Name: name, RootUrl: rootUrl, SearchUrl: searchUrl, DailyDownloadLimit: dailyDownloadLimit}
}

func (s *OneSupplierSettings) GetSearchUrl() string {
	return s.RootUrl + s.SearchUrl
}
