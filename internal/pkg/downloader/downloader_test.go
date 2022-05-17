package downloader

import (
	"testing"
)

// TODO 暂不方便在其他环境进行单元测试
func TestDownloader_DownloadSub4Movie(t *testing.T) {
	//var err error
	////dirRoot := "X:\\电影\\Spiral From the Book of Saw (2021)"
	////dirRoot := "X:\\电影\\Oslo (2021)"
	////dirRoot := "X:\\电影\\The Devil All the Time (2020)"
	////dirRoot := "X:\\电影\\21座桥 (2019)"
	////dirRoot := "X:\\电影\\An Invisible Sign (2010)"
	////dirRoot := "X:\\电影\\送你一朵小红花 (2020)"
	////dirRoot := "X:\\电影\\冰海陷落 (2018)"
	////dirRoot := "X:\\电影\\The Boss Baby Family Business (2021)"
	//dirRoot := "X:\\电影\\Don't Look Up (2021)"
	////dirRoot := "X:\\电影"
	//
	//var config = types.Config{
	//	SubNameFormatter: 0,
	//	//MovieFolder:      "X:\\电影",
	//	//SeriesFolder:     "X:\\连续剧",
	//}
	//
	//dl, err := NewDownloader(sub_formatter.GetSubFormatter(config.SubNameFormatter), types.ReqParam{
	//	SaveMultiSub:    true,
	//	SubTypePriority: 1,
	//})
	//if err != nil {
	//	t.Fatal(err)
	//}
	////err = dl.getUpdateVideoListFromEmby(config.MovieFolder, config.SeriesFolder)
	////if err != nil {
	////	t.Fatal(err)
	////}
	//err = dl.DownloadSub4Movie(dirRoot)
	//if err != nil {
	//	t.Fatal(err)
	//}
}

// TODO 暂不方便在其他环境进行单元测试
func TestDownloader_DownloadSub4Series(t *testing.T) {
	//var err error
	////dirRoot := "X:\\连续剧\\隐秘的角落 (2020)"
	////dirRoot := "X:\\连续剧\\The Bad Batch"
	//dirRoot := "X:\\连续剧\\Loki"
	////dirRoot := "X:\\连续剧\\豪斯医生 (2004)"
	////dirRoot := "X:\\连续剧\\Why Women Kill"
	////dirRoot := "X:\\连续剧\\Mare of Easttown"
	////dirRoot := "X:\\连续剧\\瑞克和莫蒂 (2013)"
	////dirRoot := "X:\\连续剧\\黑钱胜地 (2017)"
	////dirRoot := "X:\\连续剧\\黑道家族 (1999)"
	////dirRoot := "X:\\连续剧\\黑镜 (2011)"
	////dirRoot := "X:\\连续剧\\黄石 (2018)"
	////dirRoot := "X:\\连续剧\\少年间谍 (2020)"
	//
	//var config = types.Config{
	//	SubNameFormatter: 0,
	//}
	//
	//// 如果需要调试 Emby 一定需要 dirRoot := "X:\\连续剧"
	//dl, err := NewDownloader(sub_formatter.GetSubFormatter(config.SubNameFormatter), types.ReqParam{
	//	SaveMultiSub:    true,
	//	SubTypePriority: 0,
	//	//EmbyConfig:      config.EmbyConfig,
	//})
	//if err != nil {
	//	t.Fatal(err)
	//}
	////err = dl.getUpdateVideoListFromEmby(config.MovieFolder, config.SeriesFolder)
	////if err != nil {
	////	t.Fatal(err)
	////}
	//err = dl.DownloadSub4Series(dirRoot)
	//if err != nil {
	//	t.Fatal(err)
	//}
}

// TODO 暂不方便在其他环境进行单元测试
func TestDownloader_GetUpdateVideoListFromEmby(t *testing.T) {

	//se := settings.GetSettings()
	////se.CommonSettings.MoviePaths = append(se.CommonSettings.MoviePaths, "X:\\电影")
	//se.CommonSettings.SeriesPaths = append(se.CommonSettings.SeriesPaths, "X:\\TestSeries")
	////se.CommonSettings.SeriesPaths = append(se.CommonSettings.SeriesPaths, "X:\\连续剧")
	//se.EmbySettings.Enable = false
	//se.EmbySettings.AddressUrl = "http://123:8096"
	//se.EmbySettings.APIKey = "123"
	//se.EmbySettings.SkipWatched = true
	//se.EmbySettings.MoviePathsMapping["X:\\电影"] = "/mnt/share1/电影"
	//se.EmbySettings.MoviePathsMapping["X:\\连续剧"] = "/mnt/share1/连续剧"
	//
	//subSupplierHub := subSupplier.NewSubSupplierHub(
	//	*settings.GetSettings(),
	//	shooter.NewSupplier(*settings.GetSettings()),
	//)
	//
	//var err error
	//dl, err := NewDownloader(subSupplierHub, sub_formatter.GetSubFormatter(0), *se)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//err = dl.getUpdateVideoListFromEmby()
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//err = dl.DownloadSub4Movie()
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//err = dl.DownloadSub4Series()
	//if err != nil {
	//	t.Fatal(err)
	//}
}
