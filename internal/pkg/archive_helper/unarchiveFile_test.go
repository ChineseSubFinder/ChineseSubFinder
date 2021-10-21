package archive_helper

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

func TestUnArchiveFile(t *testing.T) {

	testDataPath := "../../../TestData/zips"
	testRootDir, err := pkg.CopyTestData(testDataPath)
	if err != nil {
		t.Fatal(err)
	}
	tetUnArchive(t, testRootDir, "zip.zip")
	tetUnArchive(t, testRootDir, "tar.tar")
	tetUnArchive(t, testRootDir, "rar.rar")
	tetUnArchive(t, testRootDir, "7z.7z")
}

func tetUnArchive(t *testing.T, testRootDir string, missionName string) {
	fileFPath := filepath.Join(testRootDir, missionName)
	desPath := filepath.Join(testRootDir, strings.ReplaceAll(filepath.Ext(missionName), ".", ""))
	err := UnArchiveFile(fileFPath, desPath)
	if err != nil {
		t.Fatal(err)
	}
	if pkg.IsFile(filepath.Join(desPath, subASS)) == false {
		t.Fatal(missionName, " unArchive failed")
	}
	if pkg.IsFile(filepath.Join(desPath, subSRT)) == false {
		t.Fatal(missionName, " unArchive failed")
	}
}

const subASS = "oslo.2021.1080p.web.h264-naisu.繁体&英文.ass"
const subSRT = "oslo.2021.1080p.web.h264-naisu.繁体&英文.srt"
