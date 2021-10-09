package sub_timeline_fixer

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_formatter"
	"testing"
)

func TestSubTimelineFixerHelper_FixRecentlyItemsSubTimeline(t *testing.T) {
	config := pkg.GetConfig()
	fixer := NewSubTimelineFixerHelper(config.EmbyConfig, sub_formatter.GetSubFormatter(config.SubNameFormatter))
	err := fixer.FixRecentlyItemsSubTimeline()
	if err != nil {
		t.Fatal(err)
	}
}

func TestSubTimelineFixerHelper_fixOneVideoSub(t *testing.T) {
	// What If  - S01E09    171499
	// Dan Brown's The Lost Symbol - S01E01    172412
	// 基地 S01E03 166840
	config := pkg.GetConfig()
	fixer := NewSubTimelineFixerHelper(config.EmbyConfig, sub_formatter.GetSubFormatter(config.SubNameFormatter))
	err := fixer.fixOneVideoSub("166840")
	if err != nil {
		t.Fatal(err)
	}
}
