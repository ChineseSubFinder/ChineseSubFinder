package archive_helper

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"path/filepath"
	"testing"
)

func TestUnArchiveFile(t *testing.T) {

	testDataPath := filepath.FromSlash("../../../TestData/archive_helper")
	// TODO: remove CopyTestData
	testUnArchive(t, testDataPath, "zip.zip")
	testUnArchive(t, testDataPath, "tar.tar")
	testUnArchive(t, testDataPath, "rar.rar")
	testUnArchive(t, testDataPath, "7z.7z")
}

func testUnArchive(t *testing.T, testRootDir string, missionName string) {
	fileFPath := filepath.Join(testRootDir, missionName)
	desPath := t.TempDir()
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
