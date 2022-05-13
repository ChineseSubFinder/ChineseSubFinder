package zimuku

import (
	"github.com/allanpk716/ChineseSubFinder/internal/logic/file_downloader"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/series_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/cache_center"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/rod_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/unit_test_helper"
	"path/filepath"
	"testing"
)

func TestSupplier_GetSubListFromKeyword(t *testing.T) {

	browser, err := rod_helper.NewBrowser(log_helper.GetLogger4Tester(), "", true, settings.NewSettings().AdvancedSettings.SuppliersSettings.Zimuku.RootUrl)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = browser.Close()
	}()

	//imdbId1 := "tt3228774"
	videoName := "黑白魔女库伊拉"
	s := NewSupplier(file_downloader.NewFileDownloader(cache_center.NewCacheCenter("test", settings.NewSettings(), log_helper.GetLogger4Tester())))
	outList, err := s.getSubListFromKeyword(browser, videoName)
	if err != nil {
		t.Error(err)
	}
	println(outList)
	for i, sublist := range outList {
		println(i, sublist.Name, sublist.Ext, sublist.Language.String(), sublist.Score, len(sublist.Data))
	}
}

func TestSupplier_GetSubListFromFile(t *testing.T) {

	browser, err := rod_helper.NewBrowser(log_helper.GetLogger4Tester(), "", true, settings.NewSettings().AdvancedSettings.SuppliersSettings.Zimuku.RootUrl)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = browser.Close()
	}()

	//movie1 := "X:\\电影\\The Devil All the Time (2020)\\The Devil All the Time (2020) WEBDL-1080p.mkv"
	//movie1 := "X:\\电影\\龙猫 (1988)\\龙猫 (1988) 1080p DTS.mkv"
	//movie1 := "X:\\电影\\消失爱人 (2016)\\消失爱人 (2016) 720p AAC.rmvb"
	//movie1 := "X:\\电影\\Spiral From the Book of Saw (2021)\\Spiral From the Book of Saw (2021) WEBDL-1080p.mkv"
	//movie1 := "X:\\电影\\机动战士Z高达：星之继承者 (2005)\\机动战士Z高达：星之继承者 (2005) 1080p TrueHD.mkv"
	//movie1 := "X:\\连续剧\\The Bad Batch\\Season 1\\The Bad Batch - S01E01 - Aftermath WEBDL-1080p.mkv"

	rootDir := unit_test_helper.GetTestDataResourceRootPath([]string{"sub_spplier"}, 5, true)
	movie1 := filepath.Join(rootDir, "zimuku", "movies", "The Devil All the Time (2020)", "The Devil All the Time (2020) WEBDL-1080p.mkv")
	s := NewSupplier(file_downloader.NewFileDownloader(cache_center.NewCacheCenter("test", settings.NewSettings(), log_helper.GetLogger4Tester())))
	outList, err := s.getSubListFromMovie(browser, movie1)
	if err != nil {
		t.Error(err)
	}
	println(outList)
	for i, sublist := range outList {
		println(i, sublist.Name, sublist.Ext, sublist.Language.String(), sublist.Score, len(sublist.Data))
	}
}

func TestSupplier_GetSubListFromFile4Series(t *testing.T) {

	//ser := "X:\\连续剧\\The Bad Batch" // tt12708542
	//ser := "X:\\连续剧\\杀死伊芙 (2018)"	// tt12708542
	//ser := "X:\\连续剧\\Money.Heist"
	//ser := "X:\\连续剧\\黄石 (2018)"

	// 可以指定几集去调试
	epsMap := make(map[int][]int, 0)
	epsMap[4] = []int{1}
	//epsMap[1] = []int{1, 2, 3}

	rootDir := unit_test_helper.GetTestDataResourceRootPath([]string{"sub_spplier"}, 5, true)
	ser := filepath.Join(rootDir, "zimuku", "series", "黄石 (2018)")
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

	s := NewSupplier(file_downloader.NewFileDownloader(cache_center.NewCacheCenter("test", settings.NewSettings(), log_helper.GetLogger4Tester())))
	outList, err := s.GetSubListFromFile4Series(seriesInfo)
	if err != nil {
		t.Fatal(err)
	}
	println(outList)
	for i, sublist := range outList {
		println(i, sublist.Name, sublist.Ext, sublist.Language.String(), sublist.Score, len(sublist.Data))
	}

	organizeSubFiles, err := sub_helper.OrganizeDlSubFiles(log_helper.GetLogger4Tester(), filepath.Base(seriesInfo.DirPath), outList)
	if err != nil {
		t.Fatal(err)
	}

	for s2, strings := range organizeSubFiles {
		println(s2, strings)
	}
}

func TestSupplier_getSubListFromKeyword(t *testing.T) {

	browser, err := rod_helper.NewBrowser(log_helper.GetLogger4Tester(), "", true, settings.NewSettings().AdvancedSettings.SuppliersSettings.Zimuku.RootUrl)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = browser.Close()
	}()

	//imdbID := "tt12708542" // 星球大战：残次品
	//imdbID := "tt7016936" // 杀死伊芙
	//imdbID := "tt2990738" // 恐怖直播
	//imdbID := "tt3032476" 	// 风骚律师
	//imdbID := "tt6468322" 	// 纸钞屋
	//imdbID := "tt15299712" // 云南虫谷
	//imdbID := "tt3626476"  // Vacation Friends (2021)
	imdbID := "tt11192306" // Superman.and.Lois
	zimuku := NewSupplier(file_downloader.NewFileDownloader(cache_center.NewCacheCenter("test", settings.NewSettings(), log_helper.GetLogger4Tester())))
	subInfos, err := zimuku.getSubListFromKeyword(browser, imdbID)
	if err != nil {
		t.Fatal(err)
	}
	for i, sublist := range subInfos {
		println(i, sublist.Name, sublist.Ext, sublist.Language.String(), sublist.Score, len(sublist.Data))
	}
}

func TestSupplier_step3(t *testing.T) {
	// 调试用，不作为单元测试的一个考核，因为可能不可控
	//dlUrl := "https://zmk.pw/dld/162150.html"
	//s := Supplier{}
	//fileName, datas, err := s.DownFile(dlUrl)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//println(fileName)
	//println(len(datas))
}

func TestSupplier_CheckAlive(t *testing.T) {

	s := NewSupplier(file_downloader.NewFileDownloader(cache_center.NewCacheCenter("test", settings.NewSettings(), log_helper.GetLogger4Tester())))
	alive, _ := s.CheckAlive()
	if alive == false {
		t.Fatal("CheckAlive == false")
	}
}
