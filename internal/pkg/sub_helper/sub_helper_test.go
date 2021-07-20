package sub_helper

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"path"
	"testing"
)

func TestIsOldVersionSubPrefixName(t *testing.T) {
	type args struct {
		subFileName string
	}
	tests := []struct {
		name  string
		args  args
		want  bool
		want1 string
		want2 string
	}{
		{name: "chs_en", args: args{subFileName: "Loki - S01E01 - Glorious Purpose WEBDL-1080p Proper.chs_en.ass"}, want: true, want1: ".chs_en.ass", want2: "Loki - S01E01 - Glorious Purpose WEBDL-1080p Proper.chinese(简英).default.ass"},
		{name: "chs[subhd]", args: args{subFileName: "Loki - S01E01 - Glorious Purpose WEBDL-1080p Proper.chs[subhd].ass"}, want: true, want1: ".chs[subhd].ass", want2: "Loki - S01E01 - Glorious Purpose WEBDL-1080p Proper.chinese(简,subhd).ass"},
		{name: "chs_en[shooter]", args: args{subFileName: "Loki - S01E01 - Glorious Purpose WEBDL-1080p Proper.chs_en[shooter].ass"}, want: true, want1: ".chs_en[shooter].ass", want2: "Loki - S01E01 - Glorious Purpose WEBDL-1080p Proper.chinese(简英,shooter).ass"},
		{name: "cht_en[xunlei]", args: args{subFileName: "Loki - S01E01 - Glorious Purpose WEBDL-1080p Proper.cht_en[xunlei].ass"}, want: true, want1: ".cht_en[xunlei].ass", want2: "Loki - S01E01 - Glorious Purpose WEBDL-1080p Proper.chinese(繁英,xunlei).ass"},
		{name: "zh", args: args{subFileName: "Loki - S01E01 - Glorious Purpose WEBDL-1080p Proper.zh.ass"}, want: false, want1: "", want2: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := IsOldVersionSubPrefixName(tt.args.subFileName)
			if got != tt.want {
				t.Errorf("IsOldVersionSubPrefixName() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("IsOldVersionSubPrefixName() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("IsOldVersionSubPrefixName() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestDeleteOneSeasonSubCacheFolder(t *testing.T) {

	testDataPath := "..\\..\\..\\TestData\\sub_helper"
	testRootDir, err := pkg.CopyTestData(testDataPath)
	if err != nil {
		t.Fatal(err)
	}
	err = DeleteOneSeasonSubCacheFolder(testRootDir)
	if err != nil {
		t.Fatal(err)
	}
	if pkg.IsDir(path.Join(testRootDir, "Sub_S1E0")) == true {
		t.Fatal("Sub_S1E0 not delete")
	}
}
