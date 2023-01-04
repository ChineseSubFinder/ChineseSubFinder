package archive_helper

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/unit_test_helper"
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
	if pkg.IsFile(filepath.Join(desPath, subASS)) == false {
		t.Fatal(missionName, " unArchive failed")
	}
	if pkg.IsFile(filepath.Join(desPath, subSRT)) == false {
		t.Fatal(missionName, " unArchive failed")
	}
}

const subASS = "oslo.2021.1080p.web.h264-naisu.繁体&英文.ass"
const subSRT = "oslo.2021.1080p.web.h264-naisu.繁体&英文.srt"

func TestUnArchiveFileEx(t *testing.T) {

	testRootDir := unit_test_helper.GetTestDataResourceRootPath([]string{"zips"}, 4, true)

	type args struct {
		fileFullPath string
		desRootPath  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "[zimuku]_0_Inside No 9_S7E1.zip", args: args{
				fileFullPath: filepath.Join(testRootDir, "[zimuku]_0_Inside No 9_S7E1.zip"),
				desRootPath:  filepath.Join(testRootDir, "[zimuku]_0_Inside No 9_S7E1"),
			}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := UnArchiveFileEx(tt.args.fileFullPath, tt.args.desRootPath); (err != nil) != tt.wantErr {
				t.Errorf("UnArchiveFileEx() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
