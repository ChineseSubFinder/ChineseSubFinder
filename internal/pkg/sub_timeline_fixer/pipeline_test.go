package sub_timeline_fixer

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/ass"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/srt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_parser_hub"
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
			baseSubFile:      "C:\\Tmp\\BL - S01E03\\英_2.ass",
			ffsubSyncSubFile: "C:\\Tmp\\BL - S01E03\\ffsubsync.ass",
			srcSubFile:       "C:\\Tmp\\BL - S01E03\\org.ass",
			srcFixedSubFile:  "C:\\Tmp\\BL - S01E03\\org-fix.ass",
		}, want: -4.1, wantErr: false},
		{name: "Rick and Morty - S05E10", args: args{
			baseSubFile:      "C:\\Tmp\\Rick and Morty - S05E10\\英_2.ass",
			ffsubSyncSubFile: "C:\\Tmp\\Rick and Morty - S05E10\\ffsubsync.ass",
			srcSubFile:       "C:\\Tmp\\Rick and Morty - S05E10\\org.ass",
			srcFixedSubFile:  "C:\\Tmp\\Rick and Morty - S05E10\\org-fix.ass",
		}, want: -4.1, wantErr: false},
		{name: "Foundation - S01E09", args: args{
			baseSubFile:      "C:\\Tmp\\Foundation - S01E09\\英_2.ass",
			ffsubSyncSubFile: "C:\\Tmp\\Foundation - S01E09\\ffsubsync.ass",
			srcSubFile:       "C:\\Tmp\\Foundation - S01E09\\org.ass",
			srcFixedSubFile:  "C:\\Tmp\\Foundation - S01E09\\org-fix.ass",
		}, want: -4.1, wantErr: false},
		{name: "mix", args: args{
			baseSubFile: "C:\\Tmp\\Rick and Morty - S05E10\\英_2.ass",
			srcSubFile:  "C:\\Tmp\\BL - S01E03\\org.ass",
		}, want: -4.1, wantErr: false},
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
			_, err = NewPipeline(DefaultMaxOffsetSeconds).FixTimeline(infoBase, infoSrc, false, tt.args.srcFixedSubFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("FitGSS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
