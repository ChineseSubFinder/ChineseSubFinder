package sub_timeline_fixer

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/config"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/unit_test_helper"
	"path/filepath"
	"testing"
)

func TestSubTimelineFixerHelperEx_Check(t *testing.T) {

	if NewSubTimelineFixerHelperEx(config.GetConfig().SubTimelineFixerConfig).Check() == false {
		t.Fatal("Need Install FFMPEG")
	}
}

func TestSubTimelineFixerHelperEx_Process(t *testing.T) {

	rootDir := unit_test_helper.GetTestDataResourceRootPath([]string{"sub_timeline_fixer"}, 4, true)
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
				videoFileFullPath: "X:\\连续剧\\Foundation (2021)\\Season 1\\Foundation (2021) - S01E09 - The First Crisis WEBDL-1080p.mkv",
				srcSubFPath:       filepath.Join(rootDir, "series", "Foundation (2021)", "Season 1", "Foundation (2021) - S01E09 - The First Crisis WEBDL-1080p.chinese(简英,zimuku).ass")},
			wantErr: false,
		},
	}

	s := NewSubTimelineFixerHelperEx(config.GetConfig().SubTimelineFixerConfig)
	s.Check()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := s.Process(tt.args.videoFileFullPath, tt.args.srcSubFPath); (err != nil) != tt.wantErr {
				t.Errorf("Process() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
