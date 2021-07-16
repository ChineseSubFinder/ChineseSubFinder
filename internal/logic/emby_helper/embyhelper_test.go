package emby_helper

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	"testing"
)

func init() {
	var err error
	configViper, err := pkg.InitConfigure()
	if err != nil {
		return
	}
	config, err = pkg.ReadConfig(configViper)
	if err != nil {
		return
	}
}

func TestEmbyHelper_GetRecentlyAddVideoList(t *testing.T) {

	em := NewEmbyHelper(config.EmbyConfig)
	movieList, seriesList, err := em.GetRecentlyAddVideoList(config.MovieFolder, config.SeriesFolder)
	if err != nil {
		t.Fatal(err)
	}

	println(len(movieList), len(seriesList))
}

var (
	config *types.Config
)

func TestEmbyHelper_RefreshEmbySubList(t *testing.T) {
	em := NewEmbyHelper(config.EmbyConfig)
	bok, err := em.RefreshEmbySubList()
	if err != nil {
		t.Fatal(err)
	}
	println(bok)
}