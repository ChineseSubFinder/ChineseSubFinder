package emby_helper

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"testing"
)

func TestEmbyHelper_GetRecentlyItems(t *testing.T) {

	em := NewEmbyHelper(pkg.GetConfig().EmbyConfig)
	items, err := em.GetRecentlyItems()
	if err != nil {
		t.Fatal(err)
	}

	for i, item := range items.Items {
		println(i, item.Name, item.SeriesName, item.Type)
	}
}

func TestEmbyHelper_GetItemsAncestors(t *testing.T) {
	em := NewEmbyHelper(pkg.GetConfig().EmbyConfig)
	items, err := em.GetItemAncestors("96564")
	if err != nil {
		t.Fatal(err)
	}

	if len(items) < 1 {
		t.Fatal("less than 1")
	}

	println(items[0].Name, items[0].Path)
}

func TestEmbyHelper_GetItemVideoInfo(t *testing.T) {
	em := NewEmbyHelper(pkg.GetConfig().EmbyConfig)
	// 95813 -- 命运夜
	// 96564 -- The Bad Batch - S01E11
	videoInfo, err := em.GetItemVideoInfo("95813")
	if err != nil {
		t.Fatal(err)
	}

	println(videoInfo.Name, videoInfo.Path)
}

func TestEmbyHelper_GetItemVideoInfoByUserId(t *testing.T) {
	em := NewEmbyHelper(pkg.GetConfig().EmbyConfig)
	// 95813 -- 命运夜
	// 96564 -- The Bad Batch - S01E11
	// 108766 -- R&M - S05E06
	videoInfo, err := em.GetItemVideoInfoByUserId("108766")
	if err != nil {
		t.Fatal(err)
	}

	println(videoInfo.Name, videoInfo.Path, "Default Sub Index:", videoInfo.GetDefaultSubIndex())
}

func TestEmbyHelper_UpdateVideoSubList(t *testing.T) {
	em := NewEmbyHelper(pkg.GetConfig().EmbyConfig)
	// 95813 -- 命运夜
	// 96564 -- The Bad Batch - S01E11
	err := em.UpdateVideoSubList("95813")
	if err != nil {
		t.Fatal(err)
	}
}
