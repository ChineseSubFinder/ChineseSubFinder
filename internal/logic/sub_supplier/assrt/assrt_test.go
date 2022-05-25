package assrt

import (
	"testing"

	"github.com/allanpk716/ChineseSubFinder/internal/logic/file_downloader"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/cache_center"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/random_auth_key"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
)

var assrtInstance *Supplier

func defInstance() {
	nowSettings := settings.NewSettings()
	nowSettings.SubtitleSources.AssrtSettings.Token = "nzKgRKtK11UjwdfVQ9k0LoF8gNSJY7jr"

	authKey := random_auth_key.AuthKey{
		BaseKey:  "9oDdLMZRAo",
		AESKey16: "H6HxGF99Twm4aefq",
		AESIv16:  "PGC4xC6TP3wtERnc",
	}

	assrtInstance = NewSupplier(file_downloader.NewFileDownloader(
		cache_center.NewCacheCenter("test", nowSettings, log_helper.GetLogger4Tester()), authKey))
}

func TestSupplier_getSubListFromFile(t *testing.T) {

	//videoFPath := "X:\\电影\\失控玩家 (2021)\\失控玩家 (2021).mp4"
	//isMovie := true
	defInstance()
	videoFPath := "X:\\连续剧\\杀死伊芙 (2018)\\Season 4\\Killing Eve - S04E08 - Hello, Losers WEBDL-1080p.mkv"
	//videoFPath := "X:\\连续剧\\风骚律师 (2015)\\Season 6\\Better Call Saul - S06E05 - Black and Blue WEBDL-1080p.mkv"
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
