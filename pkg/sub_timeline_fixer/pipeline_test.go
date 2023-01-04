package sub_timeline_fixer

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/log_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_parser/ass"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_parser/srt"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_parser_hub"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/unit_test_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/vad"
)

func TestPipeline_getFramerateRatios2Try(t *testing.T) {

	outList := NewPipeline(DefaultMaxOffsetSeconds).getFramerateRatios2Try()
	for i, value := range outList {
		println(i, fmt.Sprintf("%v", value))
	}
}

func TestPipeline_FitGSS(t *testing.T) {

	log := log_helper.GetLogger4Tester()
	dirRoot := unit_test_helper.GetTestDataResourceRootPath([]string{"sub_timeline_fixer"}, 4, true)
	dirRoot = filepath.Join(dirRoot, "mix")
	subParserHub := sub_parser_hub.NewSubParserHub(log, ass.NewParser(log), srt.NewParser(log))

	type args struct {
		baseSubFile      string
		ffsubSyncSubFile string
		srcSubFile       string
		srcFixedSubFile  string
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		{name: "BL S01E03", args: args{
			baseSubFile:      filepath.Join(dirRoot, "BL - S01E03", "英_2.ass"),
			ffsubSyncSubFile: filepath.Join(dirRoot, "BL - S01E03", "ffsubsync.ass"),
			srcSubFile:       filepath.Join(dirRoot, "BL - S01E03", "org.ass"),
			srcFixedSubFile:  filepath.Join(dirRoot, "BL - S01E03", "org-fix.ass"),
		}, want: -4.290000, wantErr: false},
		{name: "Rick and Morty - S05E01", args: args{
			baseSubFile:      filepath.Join(dirRoot, "Rick and Morty - S05E01", "英_2.ass"),
			ffsubSyncSubFile: filepath.Join(dirRoot, "Rick and Morty - S05E01", "ffsubsync.ass"),
			srcSubFile:       filepath.Join(dirRoot, "Rick and Morty - S05E01", "org.ass"),
			srcFixedSubFile:  filepath.Join(dirRoot, "Rick and Morty - S05E01", "org-fix.ass"),
		}, want: -6.170000, wantErr: false},
		{name: "Rick and Morty - S05E10", args: args{
			baseSubFile:      filepath.Join(dirRoot, "Rick and Morty - S05E10", "英_2.ass"),
			ffsubSyncSubFile: filepath.Join(dirRoot, "Rick and Morty - S05E10", "ffsubsync.ass"),
			srcSubFile:       filepath.Join(dirRoot, "Rick and Morty - S05E10", "org.ass"),
			srcFixedSubFile:  filepath.Join(dirRoot, "Rick and Morty - S05E10", "org-fix.ass"),
		}, want: -6.020000, wantErr: false},
		{name: "Foundation - S01E09", args: args{
			baseSubFile:      filepath.Join(dirRoot, "Foundation - S01E09", "英_2.ass"),
			ffsubSyncSubFile: filepath.Join(dirRoot, "Foundation - S01E09", "ffsubsync.ass"),
			srcSubFile:       filepath.Join(dirRoot, "Foundation - S01E09", "org.ass"),
			srcFixedSubFile:  filepath.Join(dirRoot, "Foundation - S01E09", "org-fix.ass"),
		}, want: -29.890000, wantErr: false},
		{name: "Yellowstone S04E05", args: args{
			baseSubFile:      filepath.Join(dirRoot, "Yellowstone S04E05", "英_2.ass"),
			ffsubSyncSubFile: filepath.Join(dirRoot, "Yellowstone S04E05", "ffsubsync.ass"),
			srcSubFile:       filepath.Join(dirRoot, "Yellowstone S04E05", "org.ass"),
			srcFixedSubFile:  filepath.Join(dirRoot, "Yellowstone S04E05", "org-fix.ass"),
		}, want: -62.84, wantErr: false},
		//{name: "Yellowstone S04E06", args: args{
		//	baseSubFile:      filepath.Join(dirRoot, "Yellowstone S04E06", "英_2.ass"),
		//	ffsubSyncSubFile: filepath.Join(dirRoot, "Yellowstone S04E06", "ffsubsync.ass"),
		//	srcSubFile:       filepath.Join(dirRoot, "Yellowstone S04E06", "org.ass"),
		//	srcFixedSubFile:  filepath.Join(dirRoot, "Yellowstone S04E06", "org-fix.ass"),
		//}, want: -62.84, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			bFind, infoBase, err := subParserHub.DetermineFileTypeFromFile(tt.args.baseSubFile)
			if err != nil {
				t.Fatal(err)
			}
			if bFind == false {
				t.Fatal("sub not match")
			}

			bFind, infoSrc, err := subParserHub.DetermineFileTypeFromFile(tt.args.srcSubFile)
			if err != nil {
				t.Fatal(err)
			}
			if bFind == false {
				t.Fatal("sub not match")
			}
			// ---------------------------------------------------------------------------------------
			p := NewPipeline(DefaultMaxOffsetSeconds)
			pipeResult, err := p.CalcOffsetTime(infoBase, infoSrc, nil, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("FitGSS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if pipeResult.GetOffsetTime() != tt.want {
				t.Errorf("FitGSS() offsetTome = %v, want %v", pipeResult.GetOffsetTime(), tt.want)
			}

			_, err = p.FixSubFileTimeline(infoSrc, pipeResult.ScaledFileInfo,
				pipeResult.GetOffsetTime(),
				tt.args.srcFixedSubFile)

			println(fmt.Sprintf("Offset: %f, Score: %f, Scale:%f",
				pipeResult.GetOffsetTime(),
				pipeResult.Score,
				pipeResult.ScaleFactor))
		})
	}
}

func TestPipeline_FitGSSByAudio(t *testing.T) {

	dirRoot := unit_test_helper.GetTestDataResourceRootPath([]string{"sub_timeline_fixer"}, 4, true)
	dirRoot = filepath.Join(dirRoot, "mix")

	log := log_helper.GetLogger4Tester()
	subParserHub := sub_parser_hub.NewSubParserHub(log, ass.NewParser(log), srt.NewParser(log))

	type args struct {
		audioInfo       vad.AudioInfo
		subFilePath     string
		srcFixedSubFile string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		want1   float64
		want2   float64
		wantErr bool
	}{
		// Rick and Morty - S05E01
		{name: "Rick and Morty - S05E01 -- 0",
			args: args{
				audioInfo: vad.AudioInfo{
					FileFullPath: filepath.Join(dirRoot, "Rick and Morty - S05E01", "未知语言_1.pcm"),
				},
				subFilePath:     filepath.Join(dirRoot, "Rick and Morty - S05E01", "英_2.ass"),
				srcFixedSubFile: filepath.Join(dirRoot, "Rick and Morty - S05E01", "org-fix.ass")},
			want: true, want1: 0.33,
		},
		// Rick and Morty - S05E01
		{name: "Rick and Morty - S05E01 -- 1",
			args: args{
				audioInfo: vad.AudioInfo{
					FileFullPath: filepath.Join(dirRoot, "Rick and Morty - S05E01", "未知语言_1.pcm"),
				},
				subFilePath:     filepath.Join(dirRoot, "Rick and Morty - S05E01", "org.ass"),
				srcFixedSubFile: filepath.Join(dirRoot, "Rick and Morty - S05E01", "org-fix.ass")},
			want: true, want1: -6.1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			audioVADInfos, err := vad.GetVADInfoFromAudio(vad.AudioInfo{
				FileFullPath: tt.args.audioInfo.FileFullPath,
				SampleRate:   16000,
				BitDepth:     16,
			}, true)
			if err != nil {
				t.Fatal(err)
			}

			bFind, infoSrc, err := subParserHub.DetermineFileTypeFromFile(tt.args.subFilePath)
			if err != nil {
				t.Fatal(err)
			}
			if bFind == false {
				t.Fatal("sub not match")
			}
			// ---------------------------------------------------------------------------------------
			p := NewPipeline(DefaultMaxOffsetSeconds)
			pipeResult, err := p.CalcOffsetTime(nil, infoSrc, audioVADInfos, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("FitGSS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if pipeResult.GetOffsetTime() != tt.want1 {
				t.Errorf("FitGSS() offsetTome = %v, want1 %v", pipeResult.GetOffsetTime(), tt.want1)
			}

			_, err = p.FixSubFileTimeline(infoSrc, pipeResult.ScaledFileInfo,
				pipeResult.GetOffsetTime(),
				tt.args.srcFixedSubFile)

			println(fmt.Sprintf("Offset: %f, Score: %f, Scale:%f",
				pipeResult.GetOffsetTime(),
				pipeResult.Score,
				pipeResult.ScaleFactor))
		})
	}
}
