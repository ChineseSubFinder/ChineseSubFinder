package archive_helper

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/unit_test_helper"
	"path/filepath"
	"strings"
	"testing"
)

func TestUnArchiveFile(t *testing.T) {

	testRootDir := unit_test_helper.GetTestDataResourceRootPath([]string{"zips"}, 4, true)

	testUnArchive(t, testRootDir, "zip.zip")
	testUnArchive(t, testRootDir, "tar.tar")
	testUnArchive(t, testRootDir, "rar.rar")
	testUnArchive(t, testRootDir, "7z.7z")
}

func testUnArchive(t *testing.T, testRootDir string, missionName string) {
	fileFPath := filepath.Join(testRootDir, missionName)
	desPath := filepath.Join(testRootDir, strings.ReplaceAll(filepath.Ext(missionName), ".", ""))
	err := UnArchiveFile(fileFPath, desPath)
	if err != nil {
		t.Fatal(err)
	}
	if my_util.IsFile(filepath.Join(desPath, subASS)) == false {
		t.Fatal(missionName, " unArchive failed")
	}
	if my_util.IsFile(filepath.Join(desPath, subSRT)) == false {
		t.Fatal(missionName, " unArchive failed")
	}
}

const subASS = "oslo.2021.1080p.web.h264-naisu.繁体&英文.ass"
const subSRT = "oslo.2021.1080p.web.h264-naisu.繁体&英文.srt"
