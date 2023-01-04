package ffmpeg_helper

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/log_helper"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/unit_test_helper"
)

func TestGetFFMPEGInfo(t *testing.T) {

	// use small video sample form google
	// TODO: make a video with ffmpeg on each test
	// https://gist.github.com/SeunghoonBaek/f35e0fd3db80bf55c2707cae5d0f7184
	// http://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ForBiggerEscapes.mp4
	videoFile := unit_test_helper.GetTestDataResourceRootPath([]string{"ffmpeg", "org"}, 4, false)
	videoFile = filepath.Join(videoFile, "sampleVideo.mp4")
	f := NewFFMPEGHelper(log_helper.GetLogger4Tester())
	bok, ffmpegInfo, err := f.ExportFFMPEGInfo(videoFile, Audio)
	if err != nil {
		t.Fatal(err)
	}
	if bok == false {
		t.Fatal("ExportFFMPEGInfo = false")
	}

	subArgs, audioArgs := f.getAudioAndSubExportArgs(videoFile, ffmpegInfo)

	t.Logf("\n\nsubArgs: %d   audioArgs: %d\n", len(subArgs), len(audioArgs))
}

func readString(filePath string) string {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func TestParseJsonString2GetFFMPEGInfo(t *testing.T) {

	testDataPath := unit_test_helper.GetTestDataResourceRootPath([]string{"ffmpeg", "org"}, 4, false)
	type args struct {
		videoFileFullPath string
		input             string
	}
	tests := []struct {
		name         string
		args         args
		want         bool
		subsFilter   int
		audiosFilter int
		subsFull     int
		audiosFull   int
	}{
		{name: "R&M S05E10", args: args{videoFileFullPath: "123", input: readString(filepath.Join(testDataPath, "R&M S05E10-video_stream.json"))},
			want: true, subsFilter: 1, audiosFilter: 1, subsFull: 1, audiosFull: 1},
		{name: "千与千寻", args: args{videoFileFullPath: "123", input: readString(filepath.Join(testDataPath, "千与千寻-video_stream.json"))},
			want: true, subsFilter: 2, audiosFilter: 1, subsFull: 2, audiosFull: 3},
	}

	f := NewFFMPEGHelper(log_helper.GetLogger4Tester())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := f.parseJsonString2GetFFProbeInfo(tt.args.videoFileFullPath, tt.args.input)
			if got != tt.want {
				t.Errorf("parseJsonString2GetFFProbeInfo() got = %v, want %v", got, tt.want)
			}

			if len(got1.AudioInfoList) != tt.audiosFilter || len(got1.SubtitleInfoList) != tt.subsFilter {
				t.Errorf("\n\n%s    Num. Audio: %d (%d)  Num. Subtitles: %d (%d)", tt.name, len(got1.AudioInfoList), tt.audiosFilter, len(got1.SubtitleInfoList), tt.subsFilter)
				t.Fatal("parseJsonString2GetFFProbeInfo result List < 1")
			}

			if len(got2.AudioInfoList) != tt.audiosFull || len(got2.SubtitleInfoList) != tt.subsFull {
				t.Errorf("\n\n%s    Num. Audio: %d (%d)  Num. Subtitles: %d (%d)", tt.name, len(got2.AudioInfoList), tt.audiosFull, len(got2.SubtitleInfoList), tt.subsFull)
				t.Fatal("parseJsonString2GetFFProbeInfo result List < 1")
			}
		})
	}
}

func TestExportAudioArgsByTimeRange(t *testing.T) {

	// https://www.lynxstudio.com/downloads/e44/sample-wav-file-zip-encoded-44-1khz-pcm-24-stereo/
	// TODO: make a sample audio file with ffmpeg
	// TODO: remove generated audio files
	testDataPath := unit_test_helper.GetTestDataResourceRootPath([]string{"ffmpeg"}, 4, true)
	audioFullPath := filepath.Join(testDataPath, "sampleAudio.wav")
	subFullPath := filepath.Join(testDataPath, "sampleSrt.srt")
	startTimeString := "0:0:27"
	timeLeng := "28.2"

	f := NewFFMPEGHelper(log_helper.GetLogger4Tester())

	_, _, timeRange, err := f.ExportAudioAndSubArgsByTimeRange(audioFullPath, subFullPath, startTimeString, timeLeng)
	if err != nil {
		t.Logf("\n\nTime Range: %s", timeRange)
		t.Fatal(err)
	}
}

func TestGetAudioInfo(t *testing.T) {

	testDataPath := unit_test_helper.GetTestDataResourceRootPath([]string{"ffmpeg", "org"}, 4, false)
	audioFullPath := filepath.Join(testDataPath, "sampleAudio.wav")

	f := NewFFMPEGHelper(log_helper.GetLogger4Tester())
	bok, duration, err := f.ExportAudioDurationInfo(audioFullPath)
	if err != nil || bok == false {
		t.Fatal(err)
	}

	t.Logf("\n\nAudio Duration: %f\n", duration)
}

func TestVersion(t *testing.T) {

	f := FFMPEGHelper{}
	_, err := f.Version()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("\n\nGet ffmpeg/ffprobe version\n")
}

func TestExportVideoHLSAndSubByTimeRange(t *testing.T) {

	outDirPath := "C:\\Tmp\\media\\test\\hls"
	videoFPath := "C:\\Tmp\\media\\test\\Chainsaw Man - S01E02 - ARRIVAL IN TOKYO HDTV-1080p.mp4"

	subFPaths := []string{
		"C:\\Tmp\\media\\test\\Chainsaw Man - S01E02 - ARRIVAL IN TOKYO HDTV-1080p.chinese(简,csf).default.srt",
		"C:\\Tmp\\media\\test\\Chainsaw Man - S01E01 - DOG & CHAINSAW WEBRip-1080p.chinese(简,csf).default.srt",
	}

	f := NewFFMPEGHelper(log_helper.GetLogger4Tester())
	println("Start:", time.Now().Format("2006-01-02 15:04:05"))
	m3u8, subs, err := f.ExportVideoHLSAndSubByTimeRange(videoFPath, subFPaths, "10", "10", "5.000", outDirPath)
	if err != nil {
		t.Fatal(err)
	}
	println("Start:", time.Now().Format("2006-01-02 15:04:05"))
	if pkg.IsFile(filepath.Join(outDirPath, m3u8)) == false {
		t.Fatal("m3u8 file not found")
	}

	for _, path := range subs {
		if pkg.IsFile(filepath.Join(outDirPath, path)) == false {
			t.Fatal("sub file not found")
		}
	}

}
