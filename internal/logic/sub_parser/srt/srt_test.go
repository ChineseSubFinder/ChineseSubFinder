package srt

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/types/language"
	"path/filepath"
	"testing"
)

func TestParser_DetermineFileType(t *testing.T) {
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
		{name: "1", args: args{filePath: filepath.Join(testRootDir, "[zimuku]_5_Loki.S01E02.The.Variant.1080p.DSNP.WEB-DL.DDP5.1.Atmos.H.264-CM.chs&eng.srt")}, wantNil: false, wantErr: false, wantLang: language.ChineseSimpleEnglish},
		{name: "2", args: args{filePath: filepath.Join(testRootDir, "[zimuku]_5_Loki.S01E03.Lamentis.1080p.DSNP.WEB-DL.DDP5.1.H.264-TOMMY.chs&eng.srt")}, wantNil: false, wantErr: false, wantLang: language.ChineseSimpleEnglish},
		{name: "3", args: args{filePath: filepath.Join(testRootDir, "Bridge of Spies (2015) (1080p BluRay x265 Silence).zh-cn.srt")}, wantNil: false, wantErr: false, wantLang: language.ChineseSimpleEnglish},
		{name: "4", args: args{filePath: filepath.Join(testRootDir, "Resident Evil Welcome to Raccoon City (2021) WEBRip-1080p.1.zh-cn.srt")}, wantNil: false, wantErr: false, wantLang: language.ChineseSimple},
		// 特殊一点的字幕
		// 这一个不确定是什么类型的字幕
		//{name: "4", args: args{filePath: filepath.Join(testRootDir, "SP-Empire.Of.Dreams.The.Story.Of.The.Star.Wars.Trilogy.2004.1080p.BluRay.x264.AAC5.1-[YTS.MX].zh-cn.srt")}, wantNil: false, wantErr: false, wantLang: types.ChineseSimple},
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
