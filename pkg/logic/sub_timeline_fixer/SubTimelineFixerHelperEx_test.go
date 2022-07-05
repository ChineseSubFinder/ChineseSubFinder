package sub_timeline_fixer

import (
	"testing"

	"github.com/allanpk716/ChineseSubFinder/pkg/log_helper"

	"github.com/allanpk716/ChineseSubFinder/pkg/settings"
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
