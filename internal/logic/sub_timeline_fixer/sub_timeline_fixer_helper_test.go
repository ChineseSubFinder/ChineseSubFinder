package sub_timeline_fixer

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"testing"
)

func TestSubTimelineFixerHelper_FixRecentlyItemsSubTimeline(t *testing.T) {
	config := pkg.GetConfig()
	fixer := NewSubTimelineFixerHelper(config.EmbyConfig)
	err := fixer.FixRecentlyItemsSubTimeline(config.MovieFolder, config.SeriesFolder)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSubTimelineFixerHelper_fixOneVideoSub(t *testing.T) {
	// What If  - S01E09    171499
	// Dan Brown's The Lost Symbol - S01E01    172412
	// 基地 S01E03 166840
	// 基地 S01E04 173354
	// 81873 -- R&M - S05E01
	// 145499 -- R&M - S05E10
	config := pkg.GetConfig()
	fixer := NewSubTimelineFixerHelper(config.EmbyConfig)
	err := fixer.fixOneVideoSub("173354", "")
	if err != nil {
		t.Fatal(err)
	}
}
