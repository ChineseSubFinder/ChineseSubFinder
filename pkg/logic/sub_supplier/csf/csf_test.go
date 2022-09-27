package csf

import (
	"testing"

	"github.com/allanpk716/ChineseSubFinder/pkg"

	"github.com/allanpk716/ChineseSubFinder/pkg/cache_center"
	"github.com/allanpk716/ChineseSubFinder/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/pkg/logic/file_downloader"
	"github.com/allanpk716/ChineseSubFinder/pkg/logic/series_helper"
	"github.com/allanpk716/ChineseSubFinder/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/pkg/random_auth_key"
	"github.com/allanpk716/ChineseSubFinder/pkg/settings"
)

func TestSupplier_GetSubListFromFile4Movie(t *testing.T) {

	//rootDir := unit_test_helper.GetTestDataResourceRootPath([]string{"sub_spplier"}, 5, true)
	//movie1 := filepath.Join(rootDir, "zimuku", "movies", "The Devil All the Time (2020)", "The Devil All the Time (2020) WEBDL-1080p.mkv")
	movie1 := "X:\\电影\\The Unbearable Weight of Massive Talent (2022)\\The Unbearable Weight of Massive Talent (2022) WEBRip-1080p.mkv"

	defInstance()
	got, err := csfInstance.GetSubListFromFile4Movie(movie1)
	if err != nil {
		t.Fatal(err)
	}

	for i, info := range got {
		println(i, info.FromWhere, info.Ext, info.Language.String(), len(info.Data), info.Name)
	}
}

func TestSupplier_GetSubListFromFile4Series(t *testing.T) {

	// 可以指定几集去调试
	epsMap := make(map[int][]int, 0)
	//epsMap[4] = []int{1}
	epsMap[1] = []int{1}
	//epsMap[1] = []int{1, 2, 3}

	//rootDir := unit_test_helper.GetTestDataResourceRootPath([]string{"sub_spplier"}, 5, true)
	//ser := filepath.Join(rootDir, "zimuku", "series", "黄石 (2018)")
	ser := "X:\\连续剧\\Tokyo Vice"
	// 读取本地的视频和字幕信息
	seriesInfo, err := series_helper.ReadSeriesInfoFromDir(log_helper.GetLogger4Tester(),
		ser,
		90,
		false,
		false,
		epsMap)
	if err != nil {
		t.Fatal(err)
	}

	defInstance()
	got, err := csfInstance.GetSubListFromFile4Series(seriesInfo)
	if err != nil {
		t.Fatal(err)
	}

	for i, info := range got {
		println(i, info.FromWhere, info.Ext, info.Language.String(), len(info.Data), info.Name)
	}
}

var csfInstance *Supplier

func defInstance() {

	my_util.ReadCustomAuthFile(log_helper.GetLogger4Tester())

	authKey := random_auth_key.AuthKey{
		BaseKey:  pkg.BaseKey(),
		AESKey16: pkg.AESKey16(),
		AESIv16:  pkg.AESIv16(),
	}

	nowSettings := settings.GetSettings()
	nowSettings.ExperimentalFunction.ShareSubSettings.ShareSubEnabled = true

	csfInstance = NewSupplier(file_downloader.NewFileDownloader(
		cache_center.NewCacheCenter("test", nowSettings, log_helper.GetLogger4Tester()), authKey))
}
