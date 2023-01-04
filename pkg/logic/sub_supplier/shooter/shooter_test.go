package shooter

import (
	"path/filepath"
	"testing"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/cache_center"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/log_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/file_downloader"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/random_auth_key"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/unit_test_helper"
)

func TestNewSupplier(t *testing.T) {

	pkg.ReadCustomAuthFile(log_helper.GetLogger4Tester())
	authKey := random_auth_key.AuthKey{
		BaseKey:  pkg.BaseKey(),
		AESKey16: pkg.AESKey16(),
		AESIv16:  pkg.AESIv16(),
	}

	//movie1 := "X:\\电影\\The Devil All the Time (2020)\\The Devil All the Time (2020) WEBDL-1080p.mkv"
	//movie1 := "X:\\电影\\龙猫 (1988)\\龙猫 (1988) 1080p DTS.mkv"
	//movie1 := "X:\\电影\\消失爱人 (2016)\\消失爱人 (2016) 720p AAC.rmvb"
	//movie1 := "X:\\电影\\机动战士Z高达：星之继承者 (2005)\\机动战士Z高达：星之继承者 (2005) 1080p TrueHD.mkv"
	//movie1 := "X:\\连续剧\\The Bad Batch\\Season 1\\The Bad Batch - S01E01 - Aftermath WEBDL-1080p.mkv"
	//movie1 := "X:\\电影\\An Invisible Sign (2010)\\An Invisible Sign (2010) 720p AAC.mp4"
	//movie1 := "X:\\连续剧\\少年间谍 (2020)\\Season 2\\Alex Rider - S02E01 - Episode One WEBDL-1080p.mkv"
	//movie1 := "X:\\连续剧\\黄石 (2018)\\Season 4\\Yellowstone (2018) - S04E05 - Under a Blanket of Red WEBDL-2160p.mkv"
	//movie1 := "X:\\连续剧\\瑞克和莫蒂 (2013)\\Season 5\\Rick and Morty - S05E09 - Forgetting Sarick Mortshall WEBRip-1080p.mkv"

	rootDir := unit_test_helper.GetTestDataResourceRootPath([]string{"sub_spplier"}, 5, true)
	rootDir = filepath.Join(rootDir, "shooter")

	gVideoFPath, err := unit_test_helper.GenerateShooterVideoFile(rootDir)
	if err != nil {
		t.Fatal(err)
	}
	shooter := NewSupplier(file_downloader.NewFileDownloader(cache_center.NewCacheCenter("test", log_helper.GetLogger4Tester()), authKey))
	outList, err := shooter.getSubListFromFile(gVideoFPath)
	if err != nil {
		t.Error(err)
	}
	println(outList)

	for i, sublist := range outList {
		println(i, sublist.Name, sublist.Ext, sublist.Language.String(), sublist.Score, sublist.FileUrl, len(sublist.Data))
	}

	alive, _ := shooter.CheckAlive()
	if alive == false {
		t.Fatal("CheckAlive == false")
	}
}
