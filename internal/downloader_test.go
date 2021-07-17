package internal

import (
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/ass"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/srt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	"testing"
)

func TestDownloader_DownloadSub4Movie(t *testing.T) {
	var err error
	//dirRoot := "X:\\电影\\Spiral From the Book of Saw (2021)"
	//dirRoot := "X:\\电影\\Oslo (2021)"
	//dirRoot := "X:\\电影\\The Devil All the Time (2020)"
	//dirRoot := "X:\\电影\\21座桥 (2019)"
	//dirRoot := "X:\\电影\\An Invisible Sign (2010)"
	//dirRoot := "X:\\电影\\送你一朵小红花 (2020)"
	//dirRoot := "X:\\电影\\冰海陷落 (2018)"
	dirRoot := "X:\\电影"
	config := pkg.GetConfig()
	dl := NewDownloader(types.ReqParam{
		SaveMultiSub:    true,
		SubTypePriority: 1,
		EmbyConfig:      config.EmbyConfig,
	})
	err = dl.GetUpdateVideoListFromEmby(config.MovieFolder, config.SeriesFolder)
	if err != nil {
		t.Fatal(err)
	}
	err = dl.DownloadSub4Movie(dirRoot)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDownloader_DownloadSub4Series(t *testing.T) {
	var err error
	//dirRoot := "X:\\连续剧\\隐秘的角落 (2020)"
	//dirRoot := "X:\\连续剧\\The Bad Batch"
	//dirRoot := "X:\\连续剧\\豪斯医生 (2004)"
	//dirRoot := "X:\\连续剧\\Why Women Kill"
	//dirRoot := "X:\\连续剧\\Mare of Easttown"
	//dirRoot := "X:\\连续剧\\瑞克和莫蒂 (2013)"
	//dirRoot := "X:\\连续剧\\黑钱胜地 (2017)"
	//dirRoot := "X:\\连续剧\\黑道家族 (1999)"
	//dirRoot := "X:\\连续剧\\黑镜 (2011)"
	//dirRoot := "X:\\连续剧\\黄石 (2018)"
	dirRoot := "X:\\连续剧"

	config := pkg.GetConfig()
	// 如果需要调试 Emby 一定需要 dirRoot := "X:\\连续剧"
	dl := NewDownloader(types.ReqParam{
		SaveMultiSub:    true,
		SubTypePriority: 1,
		EmbyConfig:      config.EmbyConfig,
	})
	err = dl.GetUpdateVideoListFromEmby(config.MovieFolder, config.SeriesFolder)
	if err != nil {
		t.Fatal(err)
	}
	err = dl.DownloadSub4Series(dirRoot)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDownloader_GetUpdateVideoListFromEmby(t *testing.T) {
	var err error
	config := pkg.GetConfig()
	dl := NewDownloader(types.ReqParam{
		SaveMultiSub:    true,
		SubTypePriority: 1,
		EmbyConfig:      config.EmbyConfig,
	})
	err = dl.GetUpdateVideoListFromEmby(config.MovieFolder, config.SeriesFolder)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDownloader_SubParserHub(t *testing.T) {
	//subFile := "X:\\连续剧\\瑞克和莫蒂 (2013)\\Season 4\\瑞克和莫蒂 - S04E01 - Rick and Morty.chs[zimuku].ass"
	//subFile := "X:\\连续剧\\瑞克和莫蒂 (2013)\\Season 1\\瑞克和莫蒂 - S01E01 - 试播集.en.ass"
	//subFile := "X:\\连续剧\\瑞克和莫蒂 (2013)\\Season 1\\瑞克和莫蒂 - S01E01 - 试播集.chs_en[zimuku].ass"
	//subFile := "X:\\连续剧\\瑞克和莫蒂 (2013)\\Season 4\\瑞克和莫蒂 - S04E01 - Rick and Morty.zh.srt"
	subFile := "X:\\连续剧\\黑钱胜地 (2017)\\Sub_S3E0\\[subhd]_0_Ozark.S03E07.iNTERNAL.720p.WEB.x264-GHOSTS.chs.eng.ass"

	subParserHub := sub_helper.NewSubParserHub(ass.NewParser(), srt.NewParser())
	subParserHub.IsSubHasChinese(subFile)
}