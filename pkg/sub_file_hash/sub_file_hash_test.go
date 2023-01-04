package sub_file_hash

import (
	"path/filepath"
	"testing"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/unit_test_helper"
)

func TestCalculate(t *testing.T) {

	rootDir := unit_test_helper.GetTestDataResourceRootPath([]string{"sub_spplier"}, 4, true)
	rootDir = filepath.Join(rootDir, "csf")

	gVideoFPath, err := unit_test_helper.GenerateCSFVideoFile(rootDir)
	if err != nil {
		t.Fatal(err)
	}

	calculate, err := Calculate(gVideoFPath)
	if err != nil {
		t.Fatal(err)
	}

	if calculate != checkHash {
		t.Fatal("Hash not the same")
	}

	//dd := "X:\\电影\\失控玩家 (2021)\\失控玩家 (2021).mp4"
	//calculate, err := Calculate(dd)
	//if err != nil {
	//	t.Fatal(err)
	//}
}
