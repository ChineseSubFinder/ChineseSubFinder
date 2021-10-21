package sub_helper

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"path"
	"testing"
)

func TestDeleteOneSeasonSubCacheFolder(t *testing.T) {

	testDataPath := "../../../TestData/sub_helper"
	testRootDir, err := pkg.CopyTestData(testDataPath)
	if err != nil {
		t.Fatal(err)
	}
	err = DeleteOneSeasonSubCacheFolder(testRootDir)
	if err != nil {
		t.Fatal(err)
	}
	if pkg.IsDir(filepath.Join(testRootDir, "Sub_S1E0")) == true {
		t.Fatal("Sub_S1E0 not delete")
	}
}
