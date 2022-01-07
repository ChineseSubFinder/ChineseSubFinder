package sub_timeline_fixer

import (
	"testing"
)

// 无需关注这个测试用例，这个方案暂时弃用
func TestSubTimelineFixerHelper_FixRecentlyItemsSubTimeline(t *testing.T) {
	// 单独执行这个，第一次是有效的，第二次，就没得效果了，原因是已经替换字幕了啊，当然就不会修正了啊。你懂的
	//config := config.GetConfig()
	//fixer := NewSubTimelineFixerHelper(config.EmbyConfig, config.SubTimelineFixerConfig)
	//err := fixer.FixRecentlyItemsSubTimeline(config.MovieFolder, config.SeriesFolder)
	//if err != nil {
	//	t.Fatal(err)
	//}
}

// 无需关注这个测试用例，这个方案暂时弃用
func TestSubTimelineFixerHelper_fixOneVideoSub(t *testing.T) {
	// What If  - S01E09    171499
	// Dan Brown's The Lost Symbol - S01E01    172412
	// 基地 S01E03 166840
	// 基地 S01E04 173354
	// 81873 -- R&M - S05E01
	// 145499 -- R&M - S05E10
	// 178071 -- The Night House
	//config := config.GetConfig()
	//fixer := NewSubTimelineFixerHelper(config.EmbyConfig, config.SubTimelineFixerConfig)
	//err := fixer.fixOneVideoSub("178071", "X:\\电影\\The Night House (2021)")
	//if err != nil {
	//	t.Fatal(err)
	//}
}
