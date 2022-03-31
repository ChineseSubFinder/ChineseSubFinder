package my_util

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/unit_test_helper"
	"path/filepath"
	"testing"
)

func TestCloseChrome(t *testing.T) {

	// BUG: will produce Logs under this dir
	CloseChrome()
}

func TestFileNameIsBDMV(t *testing.T) {

	rootDir := unit_test_helper.GetTestDataResourceRootPath([]string{"movies", "失控玩家 (2021)"}, 4, false)
	dbmvFPath := filepath.Join(rootDir, "CERTIFICATE", "id.bdmv")
	bok, fakeVideoFPath := FileNameIsBDMV(dbmvFPath)
	if bok == false {
		t.Fatal("FileNameIsBDMV error")
	}
	println(fakeVideoFPath)
}
