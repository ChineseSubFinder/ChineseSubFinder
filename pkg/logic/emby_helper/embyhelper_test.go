package emby_helper

import (
	"testing"

	"github.com/allanpk716/ChineseSubFinder/pkg/settings"
)

var ec = settings.EmbySettings{
	AddressUrl:            "http://192.168.50.252:8096",
	APIKey:                "xxxxx",
	MaxRequestVideoNumber: 100,
	MoviePathsMapping: map[string]string{
		"X:\\电影": "/mnt/share1/电影",
	},
	SeriesPathsMapping: map[string]string{
		"X:\\连续剧": "/mnt/share1/连续剧",
	},
}

// TODO 暂不方便在其他环境进行单元测试
func TestEmbyHelper_GetRecentlyAddVideoList(t *testing.T) {

	//embyConfig := settings.NewEmbySettings()
	//embyConfig.Enable = true
	//embyConfig.AddressUrl = "http://192.168.50.252:8096"
	//embyConfig.APIKey = "1"
	//embyConfig.SkipWatched = false
	//embyConfig.MaxRequestVideoNumber = 1000
	//embyConfig.MoviePathsMapping["X:\\电影"] = "/mnt/share1/电影"
	//embyConfig.MoviePathsMapping["X:\\连续剧"] = "/mnt/share1/连续剧"
	//
	//em := NewEmbyHelper(*embyConfig)
	//movieList, seriesList, err := em.GetRecentlyAddVideoListWithNoChineseSubtitle()
	//if err != nil {
	//	t.Fatal(err)
	//}

	//println(len(movieList), len(seriesList))
}

// TODO 暂不方便在其他环境进行单元测试
func TestEmbyHelper_RefreshEmbySubList(t *testing.T) {
	//config := config.GetConfig()
	//em := NewEmbyHelper(config.EmbyConfig)
	//bok, err := em.refreshEmbySubList()
	//if err != nil {
	//	t.Fatal(err)
	//}
	//println(bok)
}

// TODO 暂不方便在其他环境进行单元测试
func TestEmbyHelper_GetInternalEngSubAndExSub(t *testing.T) {
	//config := config.GetConfig()
	//em := NewEmbyHelper(config.EmbyConfig)
	//// 81873 -- R&M - S05E01
	//// R&M S05E10  2 org english, 5 简英 	145499
	//// 基地 S01E03 							166840
	//found, internalEngSub, exCh_EngSub, err := em.GetInternalEngSubAndExChineseEnglishSub("166840")
	//if err != nil {
	//	t.Fatal(err)
	//}
	//if found == false {
	//	t.Fatal("need found sub")
	//}
	//
	//println(internalEngSub[0].FileName, exCh_EngSub[0].FileName)
}

// TODO 暂不方便在其他环境进行单元测试
func TestEmbyHelper_GetPlayedItemsSubtitle(t *testing.T) {

	//em := NewEmbyHelper(ec)
	//moviePhyFPathMap, seriesPhyFPathMap, err := em.GetPlayedItemsSubtitle()
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//for videoPhyPath, subPhyPath := range moviePhyFPathMap {
	//	println(videoPhyPath, "--", subPhyPath)
	//	if my_util.IsFile(videoPhyPath) == false {
	//		t.Fatal("not found,", videoPhyPath)
	//	}
	//	if my_util.IsFile(subPhyPath) == false {
	//		t.Fatal("not found,", subPhyPath)
	//	}
	//}
	//
	//for videoPhyPath, subPhyPath := range seriesPhyFPathMap {
	//	println(videoPhyPath, "--", subPhyPath)
	//	if my_util.IsFile(videoPhyPath) == false {
	//		t.Fatal("not found,", videoPhyPath)
	//	}
	//	if my_util.IsFile(subPhyPath) == false {
	//		t.Fatal("not found,", subPhyPath)
	//	}
	//}
}

func TestEmbyHelper_IsVideoPlayed(t *testing.T) {

	////// 95813 -- 命运夜
	////// 96564 -- The Bad Batch - S01E11
	////// 108766 -- R&M - S05E06
	////// 145499 -- R&M - S05E10
	//tmpSettings := settings.NewSettings()
	//tmpSettings.EmbySettings = &ec
	//em := NewEmbyHelper(log_helper.GetLogger4Tester(), tmpSettings)
	//played, err := em.IsVideoPlayed("145499")
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//if played == false {
	//	t.Fatal("need played")
	//}
}
