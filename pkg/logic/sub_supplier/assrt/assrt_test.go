package assrt

import (
	"testing"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/file_downloader"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/cache_center"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/log_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/random_auth_key"
)

var assrtInstance *Supplier

func defInstance() {

	pkg.ReadCustomAuthFile(log_helper.GetLogger4Tester())

	authKey := random_auth_key.AuthKey{
		BaseKey:  pkg.BaseKey(),
		AESKey16: pkg.AESKey16(),
		AESIv16:  pkg.AESIv16(),
	}

	assrtInstance = NewSupplier(file_downloader.NewFileDownloader(
		cache_center.NewCacheCenter("test", log_helper.GetLogger4Tester()), authKey))
}

func TestSupplier_getSubListFromFile(t *testing.T) {

	//videoFPath := "X:\\电影\\失控玩家 (2021)\\失控玩家 (2021).mp4"
	//isMovie := true
	defInstance()
	//videoFPath := "X:\\连续剧\\杀死伊芙 (2018)\\Season 4\\Killing Eve - S04E08 - Hello, Losers WEBDL-1080p.mkv"
	//videoFPath := "X:\\连续剧\\Why Didn’t They Ask Evans!\\Season 1\\Why Didn’t They Ask Evans! - S01E01 - Episode 1 WEBRip-1080p.mp4"
	videoFPath := "X:\\连续剧\\Pantheon\\Season 1\\Pantheon - S01E03 - Reign of Winter WEBDL-1080p.mkv"
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
