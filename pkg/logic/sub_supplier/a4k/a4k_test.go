package a4k

import (
	"path/filepath"
	"testing"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/media_info_dealers"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/subtitle_best_api"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/file_downloader"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/series_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/unit_test_helper"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/cache_center"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/log_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/random_auth_key"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
)

func TestSupplier_searchKeyword(t *testing.T) {

	keyword := "Spider-Man: No Way Home 2021"
	defInstance()
	gotOutSubInfos, err := a4kInstance.searchKeyword(keyword, false)
	if err != nil {
		t.Fatal(err)
	}

	for i, searchResultItem := range gotOutSubInfos {
		println(i, searchResultItem.Title)
	}
}

func TestSupplier_GetSubListFromFile4Movie(t *testing.T) {

	videoFPath := "X:\\电影\\失控玩家 (2021)\\失控玩家 (2021).mp4"
	defInstance()

	gots, err := a4kInstance.GetSubListFromFile4Movie(videoFPath)
	if err != nil {
		t.Fatal(err)
	}
	for i, got := range gots {
		println(i, got.Name, len(got.Data), got.Ext)
	}
}

func TestSupplier_GetSubListFromFile4Series(t *testing.T) {

	epsMap := make(map[int][]int, 0)
	epsMap[4] = []int{1}
	//epsMap[1] = []int{1, 2, 3}
	rootDir := unit_test_helper.GetTestDataResourceRootPath([]string{"sub_spplier"}, 5, true)
	ser := filepath.Join(rootDir, "zimuku", "series", "黄石 (2018)")
	// 读取本地的视频和字幕信息
	seriesInfo, err := series_helper.ReadSeriesInfoFromDir(
		dealers,
		ser,
		90,

		false,
		false,
		epsMap)
	if err != nil {
		t.Fatal(err)
	}

	defInstance()

	gots, err := a4kInstance.GetSubListFromFile4Series(seriesInfo)
	if err != nil {
		t.Fatal(err)
	}
	for i, got := range gots {
		println(i, got.Name, len(got.Data), got.Ext)
	}

	organizeSubFiles, err := sub_helper.OrganizeDlSubFiles(log_helper.GetLogger4Tester(), filepath.Base(seriesInfo.DirPath), gots, false)
	if err != nil {
		t.Fatal(err)
	}
	for i, got := range organizeSubFiles {
		for j, s := range got {
			println(i, j, s)
		}
	}
}

var (
	a4kInstance *Supplier
	dealers     *media_info_dealers.Dealers
)

func defInstance() {

	settings.SetConfigRootPath(pkg.ConfigRootDirFPath())

	pkg.ReadCustomAuthFile(log_helper.GetLogger4Tester())

	authKey := random_auth_key.AuthKey{
		BaseKey:  pkg.BaseKey(),
		AESKey16: pkg.AESKey16(),
		AESIv16:  pkg.AESIv16(),
	}

	nowSettings := settings.Get()
	nowSettings.ExperimentalFunction.ShareSubSettings.ShareSubEnabled = true

	a4kInstance = NewSupplier(file_downloader.NewFileDownloader(
		cache_center.NewCacheCenter("test", log_helper.GetLogger4Tester()), authKey))

	dealers = media_info_dealers.NewDealers(log_helper.GetLogger4Tester(),
		subtitle_best_api.NewSubtitleBestApi(log_helper.GetLogger4Tester(), authKey))
}
