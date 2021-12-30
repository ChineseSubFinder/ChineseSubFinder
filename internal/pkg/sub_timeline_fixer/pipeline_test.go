package sub_timeline_fixer

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/ass"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/srt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_parser_hub"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/vad"
	"testing"
)

func TestPipeline_getFramerateRatios2Try(t *testing.T) {

	outList := NewPipeline(DefaultMaxOffsetSeconds).getFramerateRatios2Try()
	for i, value := range outList {
		println(i, fmt.Sprintf("%v", value))
	}
}

func TestPipeline_FitGSS(t *testing.T) {
	subParserHub := sub_parser_hub.NewSubParserHub(ass.NewParser(), srt.NewParser())

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
			baseSubFile:      "C:\\Tmp\\fixsubtimeline\\BL - S01E03\\英_2.ass",
			ffsubSyncSubFile: "C:\\Tmp\\fixsubtimeline\\BL - S01E03\\ffsubsync.ass",
			srcSubFile:       "C:\\Tmp\\fixsubtimeline\\BL - S01E03\\org.ass",
			srcFixedSubFile:  "C:\\Tmp\\fixsubtimeline\\BL - S01E03\\org-fix.ass",
		}, want: -4.290000, wantErr: false},
		{name: "Rick and Morty - S05E01", args: args{
			baseSubFile:      "C:\\Tmp\\fixsubtimeline\\Rick and Morty - S05E01\\英_2.ass",
			ffsubSyncSubFile: "C:\\Tmp\\fixsubtimeline\\Rick and Morty - S05E01\\ffsubsync.ass",
			srcSubFile:       "C:\\Tmp\\fixsubtimeline\\Rick and Morty - S05E01\\org.ass",
			srcFixedSubFile:  "C:\\Tmp\\fixsubtimeline\\Rick and Morty - S05E01\\org-fix.ass",
		}, want: -6.170000, wantErr: false},
		{name: "Rick and Morty - S05E10", args: args{
			baseSubFile:      "C:\\Tmp\\fixsubtimeline\\Rick and Morty - S05E10\\英_2.ass",
			ffsubSyncSubFile: "C:\\Tmp\\fixsubtimeline\\Rick and Morty - S05E10\\ffsubsync.ass",
			srcSubFile:       "C:\\Tmp\\fixsubtimeline\\Rick and Morty - S05E10\\org.ass",
			srcFixedSubFile:  "C:\\Tmp\\fixsubtimeline\\Rick and Morty - S05E10\\org-fix.ass",
		}, want: -6.020000, wantErr: false},
		{name: "Foundation - S01E09", args: args{
			baseSubFile:      "C:\\Tmp\\fixsubtimeline\\Foundation - S01E09\\英_2.ass",
			ffsubSyncSubFile: "C:\\Tmp\\fixsubtimeline\\Foundation - S01E09\\ffsubsync.ass",
			srcSubFile:       "C:\\Tmp\\fixsubtimeline\\Foundation - S01E09\\org.ass",
			srcFixedSubFile:  "C:\\Tmp\\fixsubtimeline\\Foundation - S01E09\\org-fix.ass",
		}, want: -29.890000, wantErr: false},
		{name: "Foundation - S01E09", args: args{
			baseSubFile:      "C:\\Tmp\\fixsubtimeline\\Yellowstone S04E05\\英_2.ass",
			ffsubSyncSubFile: "C:\\Tmp\\fixsubtimeline\\Yellowstone S04E05\\ffsubsync.ass",
			srcSubFile:       "C:\\Tmp\\fixsubtimeline\\Yellowstone S04E05\\org.ass",
			srcFixedSubFile:  "C:\\Tmp\\fixsubtimeline\\Yellowstone S04E05\\org-fix.ass",
		}, want: -62.84, wantErr: false},
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
	subParserHub := sub_parser_hub.NewSubParserHub(ass.NewParser(), srt.NewParser())

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
					FileFullPath: "C:\\Tmp\\fixsubtimeline\\Rick and Morty - S05E01\\未知语言_1.pcm",
				},
				subFilePath:     "C:\\Tmp\\fixsubtimeline\\Rick and Morty - S05E01\\英_2.ass",
				srcFixedSubFile: "C:\\Tmp\\fixsubtimeline\\Rick and Morty - S05E01\\org-fix.ass"},
			want: true, want1: 0.33,
		},
		// Rick and Morty - S05E01
		{name: "Rick and Morty - S05E01 -- 1",
			args: args{
				audioInfo: vad.AudioInfo{
					FileFullPath: "C:\\Tmp\\fixsubtimeline\\Rick and Morty - S05E01\\未知语言_1.pcm",
				},
				subFilePath:     "C:\\Tmp\\fixsubtimeline\\Rick and Morty - S05E01\\org.ass",
				srcFixedSubFile: "C:\\Tmp\\fixsubtimeline\\Rick and Morty - S05E01\\org-fix.ass"},
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
