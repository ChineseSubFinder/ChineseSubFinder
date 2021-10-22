package ffmpeg_helper

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestGetFFMPEGInfo(t *testing.T) {
	videoFile := "X:\\连续剧\\瑞克和莫蒂 (2013)\\Season 5\\Rick and Morty - S05E10 - Rickmurai Jack WEBRip-1080p.mkv"

	err := GetFFMPEGInfo(videoFile)
	if err != nil {
		t.Fatal(err)
	}
}

func readString(filePath string) string {
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func Test_parseJsonString2GetFFMPEGInfo(t *testing.T) {

	testDataPath := "../../../TestData/ffmpeg"
	testRootDir, err := pkg.CopyTestData(testDataPath)
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		videoFileFullPath string
		input             string
	}
	tests := []struct {
		name   string
		args   args
		want   bool
		subs   int
		audios int
	}{
		{name: "R&M S05E10", args: args{videoFileFullPath: "123", input: readString(filepath.Join(testRootDir, "R&M S05E10-video_stream.json"))},
			want: true, subs: 1, audios: 1},
		{name: "千与千寻", args: args{videoFileFullPath: "123", input: readString(filepath.Join(testRootDir, "千与千寻-video_stream.json"))},
			want: true, subs: 2, audios: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := parseJsonString2GetFFMPEGInfo(tt.args.videoFileFullPath, tt.args.input)
			if got != tt.want {
				t.Errorf("parseJsonString2GetFFMPEGInfo() got = %v, want %v", got, tt.want)
			}

			if len(got1.AudioInfoList) != tt.audios || len(got1.SubtitleInfoList) != tt.subs {
				t.Fatal("parseJsonString2GetFFMPEGInfo result List < 1")
			}
		})
	}
}
