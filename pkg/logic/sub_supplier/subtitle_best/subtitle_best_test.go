package subtitle_best

import (
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/cache_center"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/log_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/file_downloader"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/random_auth_key"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
	"testing"
)

var sbInstance *Supplier

func defInstance() {

	settings.SetConfigRootPath(pkg.ConfigRootDirFPath())

	pkg.ReadCustomAuthFile(log_helper.GetLogger4Tester())

	authKey := random_auth_key.AuthKey{
		BaseKey:  pkg.BaseKey(),
		AESKey16: pkg.AESKey16(),
		AESIv16:  pkg.AESIv16(),
	}

	sbInstance = NewSupplier(file_downloader.NewFileDownloader(
		cache_center.NewCacheCenter("test", log_helper.GetLogger4Tester()), authKey))
}

func TestSupplier_CheckAlive(t *testing.T) {

	defInstance()

	bok, speed := sbInstance.CheckAlive()
	println(bok, speed)
}

func TestSupplier_GetSubListFromFile4Movie(t *testing.T) {

	defInstance()

	subInfos, err := sbInstance.GetSubListFromFile4Movie("X:\\电影\\Avatar (2009)\\Avatar (2009) Bluray-1080p.mp4")
	if err != nil {
		t.Fatal(err)
		return
	}
	for i, subInfo := range subInfos {
		println(i, subInfo.Name, subInfo.GetUID())
	}
}

func TestSupplier_GetSubListFromFile4Series(t *testing.T) {

	defInstance()

	eps := "X:\\连续剧\\曼达洛人 (2019)\\Season 1\\曼达洛人 - S01E01 - 第1章：曼达洛人.mp4"
	subInfos, err := sbInstance.getSubListFromFile(eps, false, 1, 1)
	if err != nil {
		t.Fatal(err)
		return
	}

	for i, subInfo := range subInfos {
		println(i, subInfo.Name, subInfo.GetUID())
	}
}
