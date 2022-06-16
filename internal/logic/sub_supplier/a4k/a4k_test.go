package a4k

import (
	"github.com/allanpk716/ChineseSubFinder/internal/logic/series_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/unit_test_helper"
	"path/filepath"
	"testing"

	"github.com/allanpk716/ChineseSubFinder/internal/pkg/cache_center"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/global_value"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/random_auth_key"

	"github.com/allanpk716/ChineseSubFinder/internal/logic/file_downloader"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
)

func TestSupplier_searchKeyword(t *testing.T) {

	keyword := "绝命律师 第四季"
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
	seriesInfo, err := series_helper.ReadSeriesInfoFromDir(log_helper.GetLogger4Tester(), ser, 90, false, false, epsMap)
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
		println(i, got)
	}
}

var a4kInstance *Supplier

func defInstance() {

	my_util.ReadCustomAuthFile(log_helper.GetLogger4Tester())

	authKey := random_auth_key.AuthKey{
		BaseKey:  global_value.BaseKey(),
		AESKey16: global_value.AESKey16(),
		AESIv16:  global_value.AESIv16(),
	}

	nowSettings := settings.GetSettings()
	nowSettings.ExperimentalFunction.ShareSubSettings.ShareSubEnabled = true

	a4kInstance = NewSupplier(file_downloader.NewFileDownloader(
		cache_center.NewCacheCenter("test", nowSettings, log_helper.GetLogger4Tester()), authKey))
}
