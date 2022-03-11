package sub_parser_hub

import (
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/ass"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/srt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/unit_test_helper"
	"path/filepath"
	"testing"
)

func TestSubParserHubIsSubHasChinese(t *testing.T) {

	testRootDir := unit_test_helper.GetTestDataResourceRootPath([]string{"sub_parser", "org"}, 4, false)

	type args struct {
		filePath string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "1", args: args{filePath: filepath.Join(testRootDir, "[xunlei]_0_C3A5CUsers5CAdministrator5CDesktop5CThe Boss Baby Family Business_S0E0.ass")}, want: true},
		{name: "2", args: args{filePath: filepath.Join(testRootDir, "Loki - S01E01 - Glorious Purpose WEBDL-1080p Proper.chs[subhd].ass")}, want: true},
		{name: "3", args: args{filePath: filepath.Join(testRootDir, "oslo.2021.1080p.web.h264-naisu.简体&英文.ass")}, want: true},
		{name: "4", args: args{filePath: filepath.Join(testRootDir, "oslo.2021.1080p.web.h264-naisu.繁体&英文.ass")}, want: true},
		{name: "5", args: args{filePath: filepath.Join(testRootDir, "oslo.2021.1080p.web.h264-naisu.繁体.ass")}, want: true},
		{name: "6", args: args{filePath: filepath.Join(testRootDir, "[zimuku]_5_Loki.S01E02.The.Variant.1080p.DSNP.WEB-DL.DDP5.1.Atmos.H.264-CM.chs&eng.srt")}, want: true},
		{name: "7", args: args{filePath: filepath.Join(testRootDir, "[zimuku]_5_Loki.S01E03.Lamentis.1080p.DSNP.WEB-DL.DDP5.1.H.264-TOMMY.chs&eng.srt")}, want: true},
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

func TestIsEmbySubChineseLangStringWanted(t *testing.T) {
	type args struct {
		inLangString string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "00", args: args{inLangString: "chinese(简英,subhd)"}, want: true},
		{name: "01", args: args{inLangString: "chinese(简英,xunlei)"}, want: true},
		{name: "02", args: args{inLangString: "chi"}, want: true},
		{name: "03", args: args{inLangString: "chs"}, want: true},
		{name: "04", args: args{inLangString: "cht"}, want: true},

		{name: "05", args: args{inLangString: "zh-hans"}, want: true},
		{name: "06", args: args{inLangString: "zh-hant"}, want: true},
		{name: "07", args: args{inLangString: "zh-CN"}, want: true},
		{name: "08", args: args{inLangString: "zh-TW"}, want: true},
		{name: "09", args: args{inLangString: "zh-sg"}, want: true},
		{name: "10", args: args{inLangString: "zh-my"}, want: true},
		{name: "11", args: args{inLangString: "zh-hk"}, want: true},
		{name: "12", args: args{inLangString: "zh-mo"}, want: true},

		{name: "13", args: args{inLangString: "zh"}, want: true},
		{name: "14", args: args{inLangString: "en"}, want: false},
		{name: "15", args: args{inLangString: "ko"}, want: false},
		{name: "16", args: args{inLangString: "ja"}, want: false},

		{name: "17", args: args{inLangString: "zho"}, want: true},
		{name: "18", args: args{inLangString: "eng"}, want: false},
		{name: "19", args: args{inLangString: "kor"}, want: false},
		{name: "20", args: args{inLangString: "jpn"}, want: false},

		{name: "21", args: args{inLangString: "chi"}, want: true},
		{name: "22", args: args{inLangString: "eng"}, want: false},
		{name: "23", args: args{inLangString: "kor"}, want: false},
		{name: "24", args: args{inLangString: "jpn"}, want: false},

		// random text that should return false
		{name: "25", args: args{inLangString: "chineseeng"}, want: false},
		{name: "26", args: args{inLangString: "English"}, want: false},
		{name: "27", args: args{inLangString: "eng&chinese"}, want: false},
		{name: "28", args: args{inLangString: "cht&eng"}, want: false},
		{name: "29", args: args{inLangString: "chs&eng"}, want: false},
		{name: "30", args: args{inLangString: "chs_eng"}, want: false},
		{name: "31", args: args{inLangString: "cht_eng"}, want: false},
		{name: "32", args: args{inLangString: "chiasdachinese"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEmbySubChineseLangStringWanted(tt.args.inLangString); got != tt.want {
				t.Errorf("IsEmbySubChineseLangStringWanted() = %v, want %v", got, tt.want)
			}
		})
	}
}
