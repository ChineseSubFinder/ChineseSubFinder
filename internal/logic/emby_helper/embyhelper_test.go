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

func TestEmbyHelper_langStringOK(t *testing.T) {

	type args struct {
		inLang string
	}

	config := config.GetConfig()
	em := NewEmbyHelper(config.EmbyConfig)

	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "00", args: args{inLang: "chinese(简英,subhd)"}, want: true},
		{name: "01", args: args{inLang: "chinese(简英,xunlei)"}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := em.langStringOK(tt.args.inLang); got != tt.want {
				t.Errorf("langStringOK() = %v, want %v", got, tt.want)
			}
		})
	}
}
