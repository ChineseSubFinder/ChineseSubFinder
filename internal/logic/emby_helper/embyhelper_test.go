package emby_helper

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/config"
	"testing"
)

func TestEmbyHelper_GetRecentlyAddVideoList(t *testing.T) {
	config := config.GetConfig()
	em := NewEmbyHelper(config.EmbyConfig)
	movieList, seriesList, err := em.GetRecentlyAddVideoList(config.MovieFolder, config.SeriesFolder)
	if err != nil {
		t.Fatal(err)
	}

	println(len(movieList), len(seriesList))
}

func TestEmbyHelper_RefreshEmbySubList(t *testing.T) {
	config := config.GetConfig()
	em := NewEmbyHelper(config.EmbyConfig)
	bok, err := em.RefreshEmbySubList()
	if err != nil {
		t.Fatal(err)
	}
	println(bok)
}

func TestEmbyHelper_GetInternalEngSubAndExSub(t *testing.T) {
	config := config.GetConfig()
	em := NewEmbyHelper(config.EmbyConfig)
	// 81873 -- R&M - S05E01
	// R&M S05E10  2 org english, 5 简英 	145499
	// 基地 S01E03 							166840
	found, internalEngSub, exCh_EngSub, err := em.GetInternalEngSubAndExChineseEnglishSub("166840")
	if err != nil {
		t.Fatal(err)
	}
	if found == false {
		t.Fatal("need found sub")
	}

	println(internalEngSub[0].FileName, exCh_EngSub[0].FileName)
}
