package sub_timeline_fixer

import (
	"testing"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/log_helper"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
)

// TODO 暂不方便在其他环境进行单元测试
func TestSubTimelineFixerHelperEx_Check(t *testing.T) {

	//if NewSubTimelineFixerHelperEx(config.GetConfig().SubTimelineFixerConfig).Check() == false {
	//	t.Fatal("Need Install FFMPEG")
	//}
}

// TODO 暂不方便在其他环境进行单元测试
func TestSubTimelineFixerHelperEx_Process(t *testing.T) {

	//rootDir := unit_test_helper.GetTestDataResourceRootPath([]string{"sub_timeline_fixer"}, 4, true)
	type args struct {
		videoFileFullPath string
		srcSubFPath       string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Foundation (2021) - S01E09", args: args{
				//videoFileFullPath: "C:\\temp\\video\\瑞克和莫蒂 - S04E05 - Rick and Morty.mp4",
				videoFileFullPath: "C:\\temp\\video\\Rick and Morty - S05E01 - Mort Dinner Rick Andre WEBDL-1080p.mkv",
				srcSubFPath:       "C:\\temp\\video\\瑞克和莫蒂 - S04E05 - Rick and Morty.chinese(简英,zimuku).org.ass"}, // Score,48281
			//srcSubFPath: "C:\\temp\\video\\The Boys - S03E01 - Payback WEBRip-1080p.chinese(简英,subhd).ass"}, // Score,19796
			//srcSubFPath: "C:\\temp\\video\\Rick and Morty - S05E01 - Mort Dinner Rick Andre WEBDL-1080p.chinese(简英,fix).srt"}, // Score,2
			//srcSubFPath: "C:\\temp\\video\\Quo Vadis, Aida! (2021) Bluray-1080p.chinese(简,csf).default.srt"}, // Score,2
			wantErr: false,
		},
	}

	s := NewSubTimelineFixerHelperEx(log_helper.GetLogger4Tester(), *settings.NewTimelineFixerSettings())
	s.Check()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := s.Process(tt.args.videoFileFullPath, tt.args.srcSubFPath); (err != nil) != tt.wantErr {
				t.Errorf("Process() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSubTimelineFixerHelperEx_IsMatchBySubFile(t *testing.T) {

	videoFPath := "C:\\temp\\video\\Rick and Morty - S05E01 - Mort Dinner Rick Andre WEBDL-1080p.mkv"
	NowTargetSubFPath := "X:\\连续剧\\瑞克和莫蒂 (2013)\\Season 5\\Rick and Morty - S05E01 - Mort Dinner Rick Andre WEBDL-1080p.chinese(简,subhd).ass"

	logger := log_helper.GetLogger4Tester()
	s := NewSubTimelineFixerHelperEx(logger, *settings.NewTimelineFixerSettings())
	bok, ffmpegInfo, audioVADInfos, infoBase, err := s.IsVideoCanExportSubtitleAndAudio(videoFPath)
	if err != nil {
		logger.Errorln("IsVideoCanExportSubtitleAndAudio", err)
		return
	}
	if bok == false {
		logger.Errorln("IsVideoCanExportSubtitleAndAudio", "bok == false")
		return
	}

	bok, matchResult, err := s.IsMatchBySubFile(
		ffmpegInfo,
		audioVADInfos,
		infoBase,
		NowTargetSubFPath,
		CompareConfig{
			MinScore:                      40000,
			OffsetRange:                   2,
			DialoguesDifferencePercentage: 0.25,
		})
	if err != nil {
		logger.Errorln("IsMatchBySubFile", err)
		return
	}

	if bok == false && matchResult == nil {
		return
	}
}
