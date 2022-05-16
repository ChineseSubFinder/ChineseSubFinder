package assrt

import (
	"github.com/allanpk716/ChineseSubFinder/internal/logic/file_downloader"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/cache_center"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"testing"
)

func TestSupplier_getSubListFromFile(t *testing.T) {

	//videoFPath := "X:\\电影\\失控玩家 (2021)\\失控玩家 (2021).mp4"
	videoFPath := "X:\\电影\\失控玩家 (2021)\\失控玩家 (2021).mp4"
	//videoFPath := "X:\\连续剧\\风骚律师 (2015)\\Season 6\\Better Call Saul - S06E05 - Black and Blue WEBDL-1080p.mkv"

	nowSettings := settings.NewSettings()
	nowSettings.SubtitleSources.AssrtSettings.Token = "xxx"
	assrtInstance := NewSupplier(file_downloader.NewFileDownloader(cache_center.NewCacheCenter("test", nowSettings, log_helper.GetLogger4Tester())))
	got, err := assrtInstance.getSubListFromFile(videoFPath, true)
	if err != nil {
		t.Error(err)
	}
	for i, info := range got {
		println(i, info.Name, info.FileUrl)
	}
}
