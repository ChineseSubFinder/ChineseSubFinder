package sub_file_hash

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/unit_test_helper"
	"path/filepath"
	"testing"
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
}
