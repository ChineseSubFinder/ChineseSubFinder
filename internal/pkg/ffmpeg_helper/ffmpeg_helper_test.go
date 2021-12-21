package ffmpeg_helper

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestGetFFMPEGInfo(t *testing.T) {
	//videoFile := "XLen:\\连续剧\\瑞克和莫蒂 (2013)\\Season 5\\Rick and Morty - S05E10 - Rickmurai Jack WEBRip-1080p.mkv"
	videoFile := "XLen:\\连续剧\\瑞克和莫蒂 (2013)\\Season 5\\Rick and Morty - S05E01 - Mort Dinner Rick Andre WEBDL-1080p.mkv"
	//videoFile := "XLen:\\TestSeries\\Blade Runner - Black Lotus\\Season 1\\Blade Runner - Black Lotus - S01E03 - The Human Condition WEBDL-1080p.mkv"
	//videoFile := "XLen:\\连续剧\\Foundation (2021)\\Season 1\\Foundation (2021) - S01E10 - The Leap WEBDL-1080p.mkv"

	f := NewFFMPEGHelper()
	bok, ffmpegInfo, err := f.GetFFMPEGInfo(videoFile, Audio)
	if err != nil {
		t.Fatal(err)
	}
	if bok == false {
		t.Fatal("GetFFMPEGInfo = false")
	}

	subArgs, audioArgs := f.getAudioAndSubExportArgs(videoFile, ffmpegInfo)
	println(len(subArgs), len(audioArgs))
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
	testRootDir, err := my_util.CopyTestData(testDataPath)
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

	f := NewFFMPEGHelper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := f.parseJsonString2GetFFProbeInfo(tt.args.videoFileFullPath, tt.args.input)
			if got != tt.want {
				t.Errorf("parseJsonString2GetFFProbeInfo() got = %v, want %v", got, tt.want)
			}

			if len(got1.AudioInfoList) != tt.audios || len(got1.SubtitleInfoList) != tt.subs {
				t.Fatal("parseJsonString2GetFFProbeInfo result List < 1")
			}
		})
	}
}

func TestFFMPEGHelper_ExportAudioArgsByTimeRange(t *testing.T) {

	audioFullPath := "C:\\Tmp\\Rick and Morty - S05E10\\英_1.pcm"
	subFullPath := "C:\\Tmp\\Rick and Morty - S05E10\\英_2.srt"
	startTimeString := "0:1:27"
	timeLeng := "28.2"

	f := NewFFMPEGHelper()

	_, _, timeRange, err := f.ExportAudioAndSubArgsByTimeRange(audioFullPath, subFullPath, startTimeString, timeLeng)
	if err != nil {
		println(timeRange)
		t.Fatal(err)
	}
}

func TestFFMPEGHelper_GetAudioInfo(t *testing.T) {

	audioFullPath := "C:\\Tmp\\Rick and Morty - S05E10\\英_1.pcm"

	f := NewFFMPEGHelper()
	bok, duration, err := f.GetAudioDurationInfo(audioFullPath)
	if err != nil || bok == false {
		t.Fatal(err)
	}

	println(duration)
}

func TestFFMPEGHelper_Version(t *testing.T) {

	f := FFMPEGHelper{}
	got, err := f.Version()
	if err != nil {
		t.Fatal(err)
	}
	println(got)
}
