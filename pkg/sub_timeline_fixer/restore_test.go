package sub_timeline_fixer

import (
	"path/filepath"
	"testing"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/log_helper"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/unit_test_helper"
)

func Test_searchBackUpSubFile(t *testing.T) {

	files, err := searchBackUpSubFile(unit_test_helper.GetTestDataResourceRootPath([]string{"sub_timeline_fixer", "org", "movies"}, 4, false))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 3 {
		t.Fatal("backup files != 3")
	}
}

func TestRestore(t *testing.T) {

	rootDir := unit_test_helper.GetTestDataResourceRootPath([]string{"sub_timeline_fixer"}, 4, true)
	movieDir := filepath.Join(rootDir, "movies")
	seriesDir := filepath.Join(rootDir, "series")
	count, err := Restore(log_helper.GetLogger4Tester(), []string{movieDir}, []string{seriesDir})
	if err != nil {
		t.Fatal(err)
	}

	if count != 4 {
		t.Fatal("Restore files != 4")
	}
}
