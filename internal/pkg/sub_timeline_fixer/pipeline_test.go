package sub_timeline_fixer

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/ass"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/srt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_parser_hub"
	"testing"
)

func TestPipeline_getFramerateRatios2Try(t *testing.T) {

	outList := NewPipeline().getFramerateRatios2Try()
	for i, value := range outList {
		println(i, fmt.Sprintf("%v", value))
	}
}

func TestPipeline_FitGSS(t *testing.T) {
	subParserHub := sub_parser_hub.NewSubParserHub(ass.NewParser(), srt.NewParser())

	type args struct {
		baseSubFile   string
		orgFixSubFile string
		srcSubFile    string
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		{name: "BL S01E03", args: args{
			baseSubFile:   "C:\\Tmp\\BL - S01E03\\英_2.ass",
			orgFixSubFile: "C:\\Tmp\\BL - S01E03\\org-fix.ass",
			srcSubFile:    "C:\\Tmp\\BL - S01E03\\org.ass",
		}, want: -4.1, wantErr: false},
		{name: "Rick and Morty - S05E10", args: args{
			baseSubFile:   "C:\\Tmp\\Rick and Morty - S05E10\\英_2.ass",
			orgFixSubFile: "C:\\Tmp\\Rick and Morty - S05E10\\org-fix.ass",
			srcSubFile:    "C:\\Tmp\\Rick and Morty - S05E10\\org.ass",
		}, want: -4.1, wantErr: false},
		{name: "Foundation - S01E09", args: args{
			baseSubFile:   "C:\\Tmp\\Foundation - S01E09\\英_2.ass",
			orgFixSubFile: "C:\\Tmp\\Foundation - S01E09\\org-fix.ass",
			srcSubFile:    "C:\\Tmp\\Foundation - S01E09\\org.ass",
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

			//bFind, orgFix, err := subParserHub.DetermineFileTypeFromFile(tt.args.orgFixSubFile)
			//if err != nil {
			//	t.Fatal(err)
			//}
			//if bFind == false {
			//	t.Fatal("sub not match")
			//}
			// ---------------------------------------------------------------------------------------
			err = NewPipeline().FitGSS(infoBase, infoSrc)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOffsetTimeV3() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
