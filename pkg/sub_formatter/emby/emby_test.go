package emby

import (
	"testing"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/language"

	subCommon "github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_formatter/common"
)

func TestFormatter_GetFormatterName(t *testing.T) {
	f := NewFormatter()
	if f.GetFormatterName() != subCommon.FormatterNameString_Emby {
		t.Errorf("GetFormatterName error")
	}
}

func TestFormatter_IsMatchThisFormat(t *testing.T) {

	const fileWithOutExt = "The Boss Baby Family Business (2021) WEBDL-1080p"

	type args struct {
		subName string
	}
	tests := []struct {
		name  string
		args  args
		want  bool
		want1 string
		want2 string
		want3 language.MyLanguage
		want4 string
	}{
		{name: "00", args: args{subName: "The Boss Baby Family Business (2021) WEBDL-1080p.chinese(简英,subhd).ass"},
			want:  true,
			want1: fileWithOutExt,
			want2: ".ass",
			want3: language.ChineseSimpleEnglish,
			want4: "subhd"},
		{name: "01", args: args{subName: "The Boss Baby Family Business (2021) WEBDL-1080p.chinese(简英,xunlei).default.ass"},
			want:  true,
			want1: fileWithOutExt,
			want2: ".default.ass",
			want3: language.ChineseSimpleEnglish,
			want4: "xunlei"},
		{name: "02", args: args{subName: "The Boss Baby Family Business (2021) WEBDL-1080p.chinese(简英,zimuku).forced.ass"},
			want:  true,
			want1: fileWithOutExt,
			want2: ".forced.ass",
			want3: language.ChineseSimpleEnglish,
			want4: "zimuku"},
		{name: "10", args: args{subName: "The Boss Baby Family Business (2021) WEBDL-1080p.chinese(简日).ass"},
			want:  true,
			want1: fileWithOutExt,
			want2: ".ass",
			want3: language.ChineseSimpleJapanese,
			want4: ""},
		{name: "11", args: args{subName: "The Boss Baby Family Business (2021) WEBDL-1080p.chinese(简).default.ass"},
			want:  true,
			want1: fileWithOutExt,
			want2: ".default.ass",
			want3: language.ChineseSimple,
			want4: ""},
		{name: "12", args: args{subName: "The Boss Baby Family Business (2021) WEBDL-1080p.chinese(繁英).forced.ass"},
			want:  true,
			want1: fileWithOutExt,
			want2: ".forced.ass",
			want3: language.ChineseTraditionalEnglish,
			want4: ""},
		{name: "13", args: args{subName: "The Boss Baby Family Business (2021) WEBDL-1080p.chinese.ass"},
			want:  false,
			want1: "",
			want2: "",
			want3: language.Unknown,
			want4: ""},
		{name: "14", args: args{subName: "../../../TestDir/TestDir2/The Boss Baby Family Business (2021) WEBDL-1080p.chinese(简英,zimuku).forced.ass"},
			want:  true,
			want1: "../../../TestDir/TestDir2/The Boss Baby Family Business (2021) WEBDL-1080p",
			want2: ".forced.ass",
			want3: language.ChineseSimpleEnglish,
			want4: "zimuku"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Formatter{}
			got, got1, got2, got3, got4 := f.IsMatchThisFormat(tt.args.subName)
			if got != tt.want {
				t.Errorf("IsMatchThisFormat() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("IsMatchThisFormat() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("IsMatchThisFormat() got2 = %v, want %v", got2, tt.want2)
			}
			if got3 != tt.want3 {
				t.Errorf("IsMatchThisFormat() got3 = %v, want %v", got3, tt.want3)
			}
			if got4 != tt.want4 {
				t.Errorf("IsMatchThisFormat() got4 = %v, want %v", got4, tt.want4)
			}
		})
	}
}

func TestFormatter_GenerateMixSubName(t *testing.T) {

	const videoFileName = "Django Unchained (2012) Bluray-1080p.mp4"
	const videoFileNamePre = "Django Unchained (2012) Bluray-1080p"

	type args struct {
		videoFileName   string
		subExt          string
		subLang         language.MyLanguage
		extraSubPreName string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
		want2 string
	}{
		{name: "zh_shooter", args: args{videoFileName: videoFileName, subExt: common.SubExtASS, subLang: language.ChineseSimple, extraSubPreName: "shooter"},
			want:  videoFileNamePre + ".chinese(简,shooter).ass",
			want1: videoFileNamePre + ".chinese(简,shooter).default.ass",
			want2: videoFileNamePre + ".chinese(简,shooter).forced.ass"},
		{name: "zh_en_shooter", args: args{videoFileName: videoFileName, subExt: common.SubExtASS, subLang: language.ChineseSimpleEnglish, extraSubPreName: "shooter"},
			want:  videoFileNamePre + ".chinese(简英,shooter).ass",
			want1: videoFileNamePre + ".chinese(简英,shooter).default.ass",
			want2: videoFileNamePre + ".chinese(简英,shooter).forced.ass"},
		{name: "zh_en", args: args{videoFileName: videoFileName, subExt: common.SubExtASS, subLang: language.ChineseSimpleEnglish, extraSubPreName: ""},
			want:  videoFileNamePre + ".chinese(简英).ass",
			want1: videoFileNamePre + ".chinese(简英).default.ass",
			want2: videoFileNamePre + ".chinese(简英).forced.ass"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Formatter{}
			got, got1, got2 := f.GenerateMixSubName(tt.args.videoFileName, tt.args.subExt, tt.args.subLang, tt.args.extraSubPreName)
			if got != tt.want {
				t.Errorf("GenerateMixSubName() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GenerateMixSubName() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("GenerateMixSubName() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}
