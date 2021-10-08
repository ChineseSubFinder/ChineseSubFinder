package sub_timeline_fixer

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"testing"
)

func TestSubTimelineFixerHelper_FixRecentlyItemsSubTimeline(t *testing.T) {
	config := pkg.GetConfig()
	fixer := NewSubTimelineFixerHelper(config.EmbyConfig)
	err := fixer.FixRecentlyItemsSubTimeline()
	if err != nil {
		t.Fatal(err)
	}
}

func TestSubTimelineFixerHelper_fixOneVideoSub(t *testing.T) {
	// What If  - S01E09    171499
	// What If  - S01E09    172412
	config := pkg.GetConfig()
	fixer := NewSubTimelineFixerHelper(config.EmbyConfig)
	err := fixer.fixOneVideoSub("171499")
	if err != nil {
		t.Fatal(err)
	}
}
