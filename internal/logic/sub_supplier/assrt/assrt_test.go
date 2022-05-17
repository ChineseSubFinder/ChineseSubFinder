package assrt

import (
	"github.com/allanpk716/ChineseSubFinder/internal/logic/file_downloader"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/cache_center"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"testing"
)

var assrtInstance *Supplier

func defInstance() {
	nowSettings := settings.NewSettings()
	nowSettings.SubtitleSources.AssrtSettings.Token = "xxxx"
	assrtInstance = NewSupplier(file_downloader.NewFileDownloader(cache_center.NewCacheCenter("test", nowSettings, log_helper.GetLogger4Tester())))
}

func TestSupplier_getSubListFromFile(t *testing.T) {

	//videoFPath := "X:\\电影\\失控玩家 (2021)\\失控玩家 (2021).mp4"
	//isMovie := true
	defInstance()
	videoFPath := "X:\\连续剧\\风骚律师 (2015)\\Season 6\\Better Call Saul - S06E05 - Black and Blue WEBDL-1080p.mkv"
	isMovie := false

	got, err := assrtInstance.getSubListFromFile(videoFPath, isMovie)
	if err != nil {
		t.Error(err)
	}
	for i, info := range got {
		println(i, info.Name, info.FileUrl)
	}
}

func TestSupplier_CheckAlive(t *testing.T) {

	defInstance()
	bok, speed := assrtInstance.CheckAlive()
	println(bok, speed)

}
