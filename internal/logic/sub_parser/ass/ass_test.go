package ass

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	"path"
	"testing"
)

func TestParser_DetermineFileTypeFromFile(t *testing.T) {

	testDataPath := "../../../../TestData/sub_parser"
	testRootDir, err := pkg.CopyTestData(testDataPath)
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
		wantLang types.Language
	}{
		{name: "1", args: args{filePath: path.Join(testRootDir, "[xunlei]_0_C3A5CUsers5CAdministrator5CDesktop5CThe Boss Baby Family Business_S0E0.ass")}, wantNil: false, wantErr: false, wantLang: types.ChineseSimpleEnglish},
		{name: "2", args: args{filePath: path.Join(testRootDir, "Loki - S01E01 - Glorious Purpose WEBDL-1080p Proper.chs[subhd].ass")}, wantNil: false, wantErr: false, wantLang: types.ChineseSimpleEnglish},
		{name: "3", args: args{filePath: path.Join(testRootDir, "oslo.2021.1080p.web.h264-naisu.简体&英文.ass")}, wantNil: false, wantErr: false, wantLang: types.ChineseSimpleEnglish},
		{name: "4", args: args{filePath: path.Join(testRootDir, "oslo.2021.1080p.web.h264-naisu.繁体&英文.ass")}, wantNil: false, wantErr: false, wantLang: types.ChineseTraditionalEnglish},
		{name: "5", args: args{filePath: path.Join(testRootDir, "oslo.2021.1080p.web.h264-naisu.繁体.ass")}, wantNil: false, wantErr: false, wantLang: types.ChineseTraditional},
		// 特殊一点的字幕
		{name: "6", args: args{filePath: path.Join(testRootDir, "SP-Antebellum.1080p.WEB-DL.DD5.1.H.264-EVO.zh.ass")}, wantNil: false, wantErr: false, wantLang: types.ChineseTraditional},
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
