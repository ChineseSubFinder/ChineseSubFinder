package sub_parser_hub

import (
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/ass"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/srt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"path"
	"testing"
)

func TestSubParserHub_IsSubHasChinese(t *testing.T) {

	testDataPath := "../../../TestData/sub_parser"
	testRootDir, err := pkg.CopyTestData(testDataPath)
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		filePath string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "1", args: args{filePath: path.Join(testRootDir, "[xunlei]_0_C3A5CUsers5CAdministrator5CDesktop5CThe Boss Baby Family Business_S0E0.ass")}, want: true},
		{name: "2", args: args{filePath: path.Join(testRootDir, "Loki - S01E01 - Glorious Purpose WEBDL-1080p Proper.chs[subhd].ass")}, want: true},
		{name: "3", args: args{filePath: path.Join(testRootDir, "oslo.2021.1080p.web.h264-naisu.简体&英文.ass")}, want: true},
		{name: "4", args: args{filePath: path.Join(testRootDir, "oslo.2021.1080p.web.h264-naisu.繁体&英文.ass")}, want: true},
		{name: "5", args: args{filePath: path.Join(testRootDir, "oslo.2021.1080p.web.h264-naisu.繁体.ass")}, want: true},
		{name: "6", args: args{filePath: path.Join(testRootDir, "[zimuku]_5_Loki.S01E02.The.Variant.1080p.DSNP.WEB-DL.DDP5.1.Atmos.H.264-CM.chs&eng.srt")}, want: true},
		{name: "7", args: args{filePath: path.Join(testRootDir, "[zimuku]_5_Loki.S01E03.Lamentis.1080p.DSNP.WEB-DL.DDP5.1.H.264-TOMMY.chs&eng.srt")}, want: true},
	}

	subParserHub := NewSubParserHub(ass.NewParser(), srt.NewParser())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := subParserHub.IsSubHasChinese(tt.args.filePath); got != tt.want {
				t.Errorf("IsSubHasChinese() = %v, want %v", got, tt.want)
			}
		})
	}
}
