package ffmpeg_helper

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestGetSubTileIndexList(t *testing.T) {
	videoFile := "X:\\连续剧\\瑞克和莫蒂 (2013)\\Season 5\\Rick and Morty - S05E10 - Rickmurai Jack WEBRip-1080p.mkv"

	err := GetSubTileIndexList(videoFile)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_parseJsonString2GetAudioAndSubs(t *testing.T) {

	testDataPath := "../../../TestData/ffmpeg"
	testRootDir, err := pkg.CopyTestData(testDataPath)
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		input string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "R&M S05E10", args: args{input: readString(filepath.Join(testRootDir, "R&M S05E10-video_stream.json"))}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := parseJsonString2GetAudioAndSubs(tt.args.input); (err != nil) != tt.wantErr {
				t.Errorf("parseJsonString2GetAudioAndSubs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func readString(filePath string) string {
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return ""
	}
	return string(bytes)
}
