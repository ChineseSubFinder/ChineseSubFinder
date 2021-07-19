package emby_helper

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"testing"
)

func TestEmbyHelper_GetRecentlyAddVideoList(t *testing.T) {
	config := pkg.GetConfig()
	em := NewEmbyHelper(config.EmbyConfig)
	movieList, seriesList, err := em.GetRecentlyAddVideoList(config.MovieFolder, config.SeriesFolder)
	if err != nil {
		t.Fatal(err)
	}

	println(len(movieList), len(seriesList))
}

func TestEmbyHelper_RefreshEmbySubList(t *testing.T) {
	config := pkg.GetConfig()
	em := NewEmbyHelper(config.EmbyConfig)
	bok, err := em.RefreshEmbySubList()
	if err != nil {
		t.Fatal(err)
	}
	println(bok)
}
