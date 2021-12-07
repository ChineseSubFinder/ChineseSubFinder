package ass

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/types/language"
	"path/filepath"
	"testing"
)

func TestParser_DetermineFileTypeFromFile(t *testing.T) {

	testDataPath := "../../../../TestData/sub_parser"
	testRootDir, err := my_util.CopyTestData(testDataPath)
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		filePath string
	}
	tests := []struct {
		name     string
		args     args
		wantNil  bool
		wantErr  bool
		wantLang language.MyLanguage
	}{
		{name: "1", args: args{filePath: filepath.Join(testRootDir, "[xunlei]_0_C3A5CUsers5CAdministrator5CDesktop5CThe Boss Baby Family Business_S0E0.ass")}, wantNil: false, wantErr: false, wantLang: language.ChineseSimpleEnglish},
		{name: "2", args: args{filePath: filepath.Join(testRootDir, "Loki - S01E01 - Glorious Purpose WEBDL-1080p Proper.chs[subhd].ass")}, wantNil: false, wantErr: false, wantLang: language.ChineseSimpleEnglish},
		{name: "3", args: args{filePath: filepath.Join(testRootDir, "oslo.2021.1080p.web.h264-naisu.简体&英文.ass")}, wantNil: false, wantErr: false, wantLang: language.ChineseSimpleEnglish},
		{name: "4", args: args{filePath: filepath.Join(testRootDir, "oslo.2021.1080p.web.h264-naisu.繁体&英文.ass")}, wantNil: false, wantErr: false, wantLang: language.ChineseTraditionalEnglish},
		{name: "5", args: args{filePath: filepath.Join(testRootDir, "oslo.2021.1080p.web.h264-naisu.繁体.ass")}, wantNil: false, wantErr: false, wantLang: language.ChineseTraditional},
		// 特殊一点的字幕
		{name: "6", args: args{filePath: filepath.Join(testRootDir, "SP-Antebellum.1080p.WEB-DL.DD5.1.H.264-EVO.zh.ass")}, wantNil: false, wantErr: false, wantLang: language.ChineseSimpleEnglish},
		{name: "7", args: args{filePath: filepath.Join(testRootDir, "SP-Gunpowder.Milkshake.2021.1080p.WEB.h264-RUMOUR[rarbg].chinese(简).ass")}, wantNil: false, wantErr: false, wantLang: language.ChineseSimpleEnglish},
		{name: "8", args: args{filePath: filepath.Join(testRootDir, "SP-One.Hundred.And.One.Dalmatians.1961.1080p.BluRay.x264.AAC5.1-[YTS.LT].zh-cn.ass")}, wantNil: false, wantErr: false, wantLang: language.ChineseSimpleEnglish},
		{name: "9", args: args{filePath: filepath.Join(testRootDir, "SP-Pirates.of.the.Caribbean.The.Curse.of.the.Black.Pearl.2003.BluRay.1080p.x265.10bit.2Audio.MNHD-FRDS.zh-cn.ssa")}, wantNil: false, wantErr: false, wantLang: language.ChineseSimpleEnglish},
		{name: "10", args: args{filePath: filepath.Join(testRootDir, "SP-Schindlers.List.1993.BluRay.1080p.x265.10bit.2Audio.MNHD-FRDS.zh-cn.ass")}, wantNil: false, wantErr: false, wantLang: language.ChineseSimpleEnglish},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Parser{}
			bFind, got, err := p.DetermineFileTypeFromFile(tt.args.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("DetermineFileTypeFromFile() error = %v, wantErr %v", err, tt.wantErr)
				t.Fatal(err)
				return
			}

			if bFind == false {
				t.Fatal("not support sub type")
			}

			if got == nil && tt.wantNil == true {

			} else if got != nil && tt.wantNil == false {
				if got.Lang != tt.wantLang {
					t.Fatal("not wanted lang")
				}
			} else {
				t.Fatal("DetermineFileTypeFromFile got:", got, "wantNil:", tt.wantNil)
			}

			println(got.Name, got.Ext, got.Lang)
		})
	}
}
