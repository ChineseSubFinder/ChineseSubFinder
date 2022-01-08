package xunlei

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/unit_test_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	"path/filepath"
	"testing"
)

func TestGetList(t *testing.T) {
	//movie1 := "X:\\电影\\The Devil All the Time (2020)\\The Devil All the Time (2020) WEBDL-1080p.mkv"
	//movie1 := "X:\\电影\\龙猫 (1988)\\龙猫 (1988) 1080p DTS.mkv"
	//movie1 := "X:\\电影\\消失爱人 (2016)\\消失爱人 (2016) 720p AAC.rmvb"
	//movie1 := "X:\\电影\\Spiral From the Book of Saw (2021)\\Spiral From the Book of Saw (2021) WEBDL-1080p.mkv"
	//movie1 := "X:\\电影\\机动战士Z高达：星之继承者 (2005)\\机动战士Z高达：星之继承者 (2005) 1080p TrueHD.mkv"
	//movie1 := "X:\\连续剧\\The Bad Batch\\Season 1\\The Bad Batch - S01E01 - Aftermath WEBDL-1080p.mkv"
	//movie1 := "X:\\连续剧\\黄石 (2018)\\Season 4\\Yellowstone (2018) - S04E05 - Under a Blanket of Red WEBDL-2160p.mkv"
	//movie1 := "X:\\动漫\\碧蓝之海 (2018)\\Season 1\\碧蓝之海 - S01E01 - [UHA-WINGS][Grand Blue][01][BDRIP 1920x1080 x264 FLACx2].mkv"
	//movie1 := "X:\\连续剧\\瑞克和莫蒂 (2013)\\Season 5\\Rick and Morty - S05E01 - Mort Dinner Rick Andre WEBDL-1080p.mkv"
	//movie1 := "X:\\电影\\手机 (2003)\\手机 (2003) 720p Cooker.rmvb"

	rootDir := unit_test_helper.GetTestDataResourceRootPath([]string{"sub_spplier"}, 5, true)
	rootDir = filepath.Join(rootDir, "xunlei")

	gVideoFPath, err := unit_test_helper.GenerateXunleiVideoFile(rootDir)
	if err != nil {
		t.Fatal(err)
	}

	xunlie := NewSupplier(types.ReqParam{Topic: 3})
	outList, err := xunlie.getSubListFromFile(gVideoFPath)
	if err != nil {
		t.Error(err)
	}
	println(outList)

	for i, sublist := range outList {
		println(i, sublist.Name, sublist.Ext, sublist.Language.String(), sublist.Score, len(sublist.Data))
	}
}
