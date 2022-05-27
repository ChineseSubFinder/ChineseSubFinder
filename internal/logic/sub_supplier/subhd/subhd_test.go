package subhd

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/allanpk716/ChineseSubFinder/internal/logic/file_downloader"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/series_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/cache_center"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/global_value"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/random_auth_key"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/something_static"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/unit_test_helper"
	commonValue "github.com/allanpk716/ChineseSubFinder/internal/types/common"
)

var authKey random_auth_key.AuthKey

func defInstance() {
	my_util.ReadCustomAuthFile(log_helper.GetLogger4Tester())
	authKey = random_auth_key.AuthKey{
		BaseKey:  global_value.BaseKey(),
		AESKey16: global_value.AESKey16(),
		AESIv16:  global_value.AESIv16(),
	}
}

// 无需关注这个测试用例，这个方案暂时弃用
func TestSupplier_GetSubListFromFile(t *testing.T) {

	//movie1 := "X:\\电影\\The Devil All the Time (2020)\\The Devil All the Time (2020) WEBDL-1080p.mkv"
	//movie1 := "X:\\电影\\Luca (2021)\\Luca (2021) WEBDL-1080p.mkv"
	//movie1 := "X:\\电影\\The Boss Baby Family Business (2021)\\The Boss Baby Family Business (2021) WEBDL-1080p.mkv"
	//movie1 := "X:\\电影\\Oslo (2021)\\Oslo (2021) WEBDL-1080p.mkv"
	//movie1 := "X:\\电影\\Spiral From the Book of Saw (2021)\\Spiral From the Book of Saw (2021) WEBDL-1080p.mkv"
	//movie1 := "X:\\电影\\消失爱人 (2016)\\消失爱人 (2016) 720p AAC.rmvb"
	//movie1 := "X:\\电影\\机动战士Z高达：星之继承者 (2005)\\机动战士Z高达：星之继承者 (2005) 1080p TrueHD.mkv"
	getCode()
	defInstance()
	rootDir := unit_test_helper.GetTestDataResourceRootPath([]string{"sub_spplier"}, 5, true)
	movie1 := filepath.Join(rootDir, "zimuku", "movies", "消失爱人 (2016)", "消失爱人 (2016) 720p AAC.rmvb")

	subhd := NewSupplier(file_downloader.NewFileDownloader(cache_center.NewCacheCenter("test", settings.NewSettings(), log_helper.GetLogger4Tester()), authKey))
	outList, err := subhd.getSubListFromFile4Movie(movie1)
	if err != nil {
		t.Error(err)
	}
	println(outList)

	if len(outList) == 0 {
		println("now sub found")
	}

	for i, sublist := range outList {
		println(i, sublist.Name, sublist.Ext, sublist.Language.String(), sublist.Score, sublist.FileUrl, len(sublist.Data))
	}

	alive, _ := subhd.CheckAlive()
	if alive == false {
		t.Fatal("CheckAlive == false")
	}
}

// 无需关注这个测试用例，这个方案暂时弃用
func TestSupplier_GetSubListFromFile4Series(t *testing.T) {

	//ser := "X:\\连续剧\\The Bad Batch"	// tt12708542
	//ser := "X:\\连续剧\\瑞克和莫蒂 (2013)" //
	//ser := "X:\\连续剧\\杀死伊芙 (2018)"	// tt7016936
	//ser := "X:\\连续剧\\Money.Heist"
	//ser := "X:\\连续剧\\黑钱胜地 (2017)"
	getCode()
	defInstance()
	// 可以指定几集去调试
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
	s := NewSupplier(file_downloader.NewFileDownloader(cache_center.NewCacheCenter("test", settings.NewSettings(), log_helper.GetLogger4Tester()), authKey))
	outList, err := s.GetSubListFromFile4Series(seriesInfo)
	if err != nil {
		t.Fatal(err)
	}
	for i, sublist := range outList {
		println(i, sublist.Name, sublist.Ext, sublist.Language.String(), sublist.Score, len(sublist.Data))
	}
}

// 无需关注这个测试用例，这个方案暂时弃用
func TestSupplier_getSubListFromKeyword4Movie(t *testing.T) {

	//imdbID := "tt12708542" // 星球大战：残次品
	//imdbID := "tt7016936" // 杀死伊芙
	imdbID := "tt2990738" // 恐怖直播
	//imdbID := "tt3032476" 	// 风骚律师
	//imdbID := "tt6468322" 	// 纸钞屋
	//imdbID := "tt15299712" // 云南虫谷
	//imdbID := "tt3626476" // Vacation Friends (2021)
	getCode()
	defInstance()
	subhd := NewSupplier(file_downloader.NewFileDownloader(cache_center.NewCacheCenter("test", settings.NewSettings(), log_helper.GetLogger4Tester()), authKey))
	subInfos, err := subhd.getSubListFromKeyword4Movie(imdbID)
	if err != nil {
		t.Fatal(err)
	}
	for i, sublist := range subInfos {
		println(i, sublist.Name, sublist.Ext, sublist.Language.String(), sublist.Score, len(sublist.Data))
	}
}

func getCode() {

	nowTT := time.Now()
	nowTimeFileNamePrix := fmt.Sprintf("%d%d%d", nowTT.Year(), nowTT.Month(), nowTT.Day())
	updateTimeString, code, err := something_static.GetCodeFromWeb(log_helper.GetLogger4Tester(), nowTimeFileNamePrix)
	if err != nil {
		commonValue.SubhdCode = ""
	} else {
		commonValue.SubhdCode = code
	}
	fmt.Println("UpdateTime", updateTimeString)
}
