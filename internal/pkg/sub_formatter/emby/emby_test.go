package emby

import (
	"github.com/allanpk716/ChineseSubFinder/internal/common"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	"testing"
)

func TestFormatter_GetFormatterName(t *testing.T) {
	f := NewFormatter()
	if f.GetFormatterName() != "emby formatter" {
		t.Errorf("GetFormatterName error")
	}
}

func TestFormatter_IsMatchThisFormat(t *testing.T) {
	type args struct {
		subName string
	}
	tests := []struct {
		name  string
		args  args
		want  bool
		want1 string
		want2 types.Language
		want3 string
	}{
		{name: "00", args: args{subName: "The Boss Baby Family Business (2021) WEBDL-1080p.chinese(简英,subhd).ass"},
			want:  true,
			want1: ".ass",
			want2: types.ChineseSimpleEnglish,
			want3: "subhd"},
		{name: "01", args: args{subName: "The Boss Baby Family Business (2021) WEBDL-1080p.chinese(简英,xunlei).default.ass"},
			want:  true,
			want1: ".default.ass",
			want2: types.ChineseSimpleEnglish,
			want3: "xunlei"},
		{name: "02", args: args{subName: "The Boss Baby Family Business (2021) WEBDL-1080p.chinese(简英,zimuku).forced.ass"},
			want:  true,
			want1: ".forced.ass",
			want2: types.ChineseSimpleEnglish,
			want3: "zimuku"},
		{name: "10", args: args{subName: "The Boss Baby Family Business (2021) WEBDL-1080p.chinese(简日).ass"},
			want:  true,
			want1: ".ass",
			want2: types.ChineseSimpleJapanese,
			want3: ""},
		{name: "11", args: args{subName: "The Boss Baby Family Business (2021) WEBDL-1080p.chinese(简).default.ass"},
			want:  true,
			want1: ".default.ass",
			want2: types.ChineseSimple,
			want3: ""},
		{name: "12", args: args{subName: "The Boss Baby Family Business (2021) WEBDL-1080p.chinese(繁英).forced.ass"},
			want:  true,
			want1: ".forced.ass",
			want2: types.ChineseTraditionalEnglish,
			want3: ""},
		{name: "03", args: args{subName: "The Boss Baby Family Business (2021) WEBDL-1080p.chinese.ass"},
			want:  false,
			want1: "",
			want2: types.ChineseSimple,
			want3: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Formatter{}
			got, got1, got2, got3 := f.IsMatchThisFormat(tt.args.subName)
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
		})
	}
}

func TestFormatter_GenerateMixSubName(t *testing.T) {

	const videoFileName = "Django Unchained (2012) Bluray-1080p.mp4"
	const videoFileNamePre = "Django Unchained (2012) Bluray-1080p"

	type args struct {
		videoFileName   string
		subExt          string
		subLang         types.Language
		extraSubPreName string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
		want2 string
	}{
		{name: "zh_shooter", args: args{videoFileName: videoFileName, subExt: common.SubExtASS, subLang: types.ChineseSimple, extraSubPreName: "shooter"},
			want:  videoFileNamePre + ".chinese(简,shooter).ass",
			want1: videoFileNamePre + ".chinese(简,shooter).default.ass",
			want2: videoFileNamePre + ".chinese(简,shooter).forced.ass"},
		{name: "zh_en_shooter", args: args{videoFileName: videoFileName, subExt: common.SubExtASS, subLang: types.ChineseSimpleEnglish, extraSubPreName: "shooter"},
			want:  videoFileNamePre + ".chinese(简英,shooter).ass",
			want1: videoFileNamePre + ".chinese(简英,shooter).default.ass",
			want2: videoFileNamePre + ".chinese(简英,shooter).forced.ass"},
		{name: "zh_en", args: args{videoFileName: videoFileName, subExt: common.SubExtASS, subLang: types.ChineseSimpleEnglish, extraSubPreName: ""},
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
