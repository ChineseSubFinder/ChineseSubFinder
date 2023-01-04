package TestCode

import (
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/log_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_timeline_fixer"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
	"path/filepath"
	"testing"
)

func Test_statistics_subs_score(t *testing.T) {
	type args struct {
		baseAudioFileFPath string
		baseSubFileFPath   string
		subSearchRootPath  string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test_statistics_subs_score",
			args: args{
				baseAudioFileFPath: "C:\\temp\\video\\base\\RM-S05E01\\未知语言_1.pcm",
				baseSubFileFPath:   "C:\\temp\\video\\base\\RM-S05E01\\英_2.srt",
				subSearchRootPath:  "X:\\电影",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statistics_subs_score(tt.args.baseAudioFileFPath, tt.args.baseSubFileFPath, tt.args.subSearchRootPath)
		})
	}
}

func Test_statistics_subs_score_one(t *testing.T) {
	type args struct {
		baseAudioFileFPath string
		baseSubFileFPath   string
		srcSubFileFPath    string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test_statistics_subs_score_one",
			args: args{
				baseAudioFileFPath: "C:\\temp\\video\\base\\RM-S05E01\\未知语言_1.pcm",
				baseSubFileFPath:   "C:\\temp\\video\\base\\RM-S05E01\\英_2.srt",
				srcSubFileFPath:    "C:\\temp\\video\\Rick and Morty - S05E01 - Mort Dinner Rick Andre WEBDL-1080p.chinese(简英,fix).srt",
			},
		},
		{
			name: "Test_statistics_subs_score_one2",
			args: args{
				baseAudioFileFPath: "C:\\temp\\video\\base\\RM-S05E01\\未知语言_1.pcm",
				baseSubFileFPath:   "C:\\temp\\video\\base\\RM-S05E01\\英_2.srt",
				srcSubFileFPath:    "C:\\temp\\video\\The Boys - S03E01 - Payback WEBRip-1080p.chinese(简英,subhd).ass",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statistics_subs_score_one(tt.args.baseAudioFileFPath, tt.args.baseSubFileFPath, tt.args.srcSubFileFPath)
		})
	}
}

func Test_statistics_subs_score_is_match(t *testing.T) {
	type args struct {
		videoFPath        string
		subSearchRootPath string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test_statistics_subs_score_is_match",
			args: args{
				videoFPath:        "C:\\temp\\video\\Rick and Morty - S05E01 - Mort Dinner Rick Andre WEBDL-1080p.mkv",
				subSearchRootPath: "X:\\连续剧\\瑞克和莫蒂 (2013)",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			logger := log_helper.GetLogger4Tester()
			s := sub_timeline_fixer.NewSubTimelineFixerHelperEx(logger, *settings.NewTimelineFixerSettings())
			bok, ffmpegInfo, audioVADInfos, infoBase, err := s.IsVideoCanExportSubtitleAndAudio(tt.args.videoFPath)
			if err != nil {
				logger.Errorln("IsVideoCanExportSubtitleAndAudio", err)
				return
			}
			if bok == false {
				logger.Errorln("IsVideoCanExportSubtitleAndAudio", "bok == false")
				return
			}

			statistics_subs_score_is_match(logger, s, ffmpegInfo, audioVADInfos, infoBase, tt.args.subSearchRootPath, filepath.Base(tt.args.videoFPath))
		})
	}
}
