package emby_api

import (
	"testing"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
)

var ec = settings.EmbySettings{
	Enable:                true,
	AddressUrl:            "http://192.168.50.252:8096",
	APIKey:                "xxxx",
	MaxRequestVideoNumber: 100,
}

// TODO 暂不方便在其他环境进行单元测试
func TestEmbyHelper_GetRecentlyItems(t *testing.T) {

	//em := NewEmbyApi(ec)
	//items, err := em.GetRecentlyItems()
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//for i, item := range items.Items {
	//	println(i, item.Name, item.SeriesName, item.Type)
	//}
}

// TODO 暂不方便在其他环境进行单元测试
func TestEmbyHelper_GetItemsAncestors(t *testing.T) {

	// 95813 -- 命运夜
	// 96564 -- The Bad Batch - S01E11
	// R&M S05E10  2 org english, 5 简英 145499
	// R&M 15430
	// 基地 S01E03 166840
	// 基地 S01E04 173354
	// 算牌人 166837
	// 327198
	//em := NewEmbyApi(&ec)
	//items, err := em.GetItemAncestors("145499")
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//if len(items) < 1 {
	//	t.Fatal("less than 1")
	//}
	//
	//println(items[0].Name, items[0].Path)
}

// TODO 暂不方便在其他环境进行单元测试
func TestEmbyHelper_GetItemVideoInfo(t *testing.T) {
	//em := NewEmbyApi(log_helper.GetLogger4Tester(), &ec)
	//// 95813 -- 命运夜
	//// 96564 -- The Bad Batch - S01E11
	//// R&M S05E10  2 org english, 5 简英 145499
	//// R&M 15430
	//// 基地 S01E03 166840
	//// 基地 S01E04 173354
	//// 算牌人 166837
	//// 327198
	//videoInfo, err := em.GetItemVideoInfo("15430")
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//println(videoInfo.Name, videoInfo.Path, videoInfo.MediaSources[0].Id)
}

// TODO 暂不方便在其他环境进行单元测试
func TestEmbyHelper_GetItemVideoInfoByUserId(t *testing.T) {
	//em := NewEmbyApi(log_helper.GetLogger4Tester(), &ec)
	//// 95813 -- 命运夜
	//// 96564 -- The Bad Batch - S01E11
	//// 108766 -- R&M - S05E06
	//// 145499 -- R&M - S05E10
	////videoInfo, err := em.GetItemVideoInfoByUserId("6a9aa3be30534e668e58640123890a7b", "145499")
	//videoInfo, err := em.GetItemVideoInfoByUserId("c248ec6305374192bdf892d4b9739f80", "145499")
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//println(videoInfo.Name, videoInfo.Path, "Default Sub OffsetIndex:", videoInfo.GetDefaultSubIndex())
}

// TODO 暂不方便在其他环境进行单元测试
func TestEmbyHelper_UpdateVideoSubList(t *testing.T) {
	//em := NewEmbyApi(log_helper.GetLogger4Tester(), &ec)
	//// 95813 -- 命运夜
	//// 96564 -- The Bad Batch - S01E11
	//// 81873 -- R&M - S05E01
	//// 145499 -- R&M - S05E10
	//// 161434 -- 基地 S01E02
	//// 166840 -- 基地 S01E03
	//// 173354 -- 基地 S01E04
	//// 172412 -- Dan Brown's The Lost Symbol S01E01
	//// 194046 -- 窃贼军团
	//// 178071 -- The Night House
	//// 215162 --  Black Lotus - S01E03
	//// 229865 --  黄石 - S04E06
	//// 433745 --  攻壳机动队 1995
	//err := em.UpdateVideoSubList("433745")
	//if err != nil {
	//	t.Fatal(err)
	//}
}

// TODO 暂不方便在其他环境进行单元测试
func TestEmbyHelper_GetUserIdList(t *testing.T) {
	//em := NewEmbyApi(&ec)
	//userIds, err := em.GetUserIdList()
	//if err != nil {
	//	t.Fatal(err)
	//}
	//for i, item := range userIds.Items {
	//	t.Logf("\n\n%d  %s  %s", i, item.Name, item.Id)
	//}
}

// TODO 暂不方便在其他环境进行单元测试
func TestEmbyApi_GetSubFileData(t *testing.T) {
	//em := NewEmbyApi(config.GetConfig().EmbyConfig)
	//// R&M S05E10  2 org english, 5 简英					"145499", "c4678509adb72a8b5034bdac2f1fccde", "5", ".ass"
	//// 基地 S01E03		2=eng 	6=chi 	45=简英			"166840", "d6c68ec6097aeceb9f5c1d82add66213", "2", ".ass"
	//// 基地 S01E04		2=eng 	6=chi 	45=简英			"173354", "c08f514cc1708f3fadea56e489da33db", "2", ".ass"
	//
	//subFileData, err := em.GetSubFileData("173354", "c08f514cc1708f3fadea56e489da33db", "3", ".ass")
	////subFileData, err := em.GetSubFileData("145499", "c4678509adb72a8b5034bdac2f1fccde", "5", ".ass")
	////subFileData, err := em.GetSubFileData("166840", "d6c68ec6097aeceb9f5c1d82add66213", "45", ".ass")
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//println(subFileData)
}

// TODO 暂不方便在其他环境进行单元测试
func TestEmbyApi_RefreshRecentlyVideoInfo(t *testing.T) {

	//em := NewEmbyApi(config.GetConfig().EmbyConfig)
	//err := em.RefreshRecentlyVideoInfo()
	//if err != nil {
	//	t.Fatal("RefreshRecentlyVideoInfo() error = " + err.Error())
	//}
}
