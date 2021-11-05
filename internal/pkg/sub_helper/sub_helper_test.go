package sub_helper

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"path/filepath"
	"testing"
)

func TestDeleteOneSeasonSubCacheFolder(t *testing.T) {

	testDataPath := "../../../TestData/sub_helper"
	testRootDir, err := my_util.CopyTestData(testDataPath)
	if err != nil {
		t.Fatal(err)
	}
	err = DeleteOneSeasonSubCacheFolder(testRootDir)
	if err != nil {
		t.Fatal(err)
	}
	if my_util.IsDir(filepath.Join(testRootDir, "Sub_S1E0")) == true {
		t.Fatal("Sub_S1E0 not delete")
	}
}

func TestSearchMatchedSubFileByOneVideo(t *testing.T) {

	testDataPath := "../../../TestData/sub_helper"
	testRootDir, err := my_util.CopyTestData(testDataPath)
	if err != nil {
		t.Fatal(err)
	}
	videoFPath := filepath.Join(testRootDir, "R&M-S05E10", "Rick and Morty - S05E10 - Rickmurai Jack WEBRip-1080p.mp4")
	subFiles, err := SearchMatchedSubFileByOneVideo(videoFPath)
	if err != nil {
		t.Fatal(err)
	}

	if len(subFiles) != 5 {
		t.Fatal("subFiles len != 5")
	}
}
