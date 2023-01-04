package srt

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/language"

	lan "github.com/ChineseSubFinder/ChineseSubFinder/pkg/language"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/log_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/unit_test_helper"
)

func TestParser_DetermineFileType(t *testing.T) {
	testRootDir := unit_test_helper.GetTestDataResourceRootPath([]string{"sub_parser"}, 5, true)
	type args struct {
		filePath string
	}
	tests := []struct {
		name            string
		args            args
		wantNil         bool
		wantErr         bool
		wantDialogueLen int
		wantLang        language.MyLanguage
	}{
		{name: "1", args: args{filePath: filepath.Join(testRootDir, "[zimuku]_5_Loki.S01E02.The.Variant.1080p.DSNP.WEB-DL.DDP5.1.Atmos.H.264-CM.chs&eng.srt")}, wantNil: false, wantErr: false, wantLang: language.ChineseSimpleEnglish, wantDialogueLen: 809},
		{name: "2", args: args{filePath: filepath.Join(testRootDir, "[zimuku]_5_Loki.S01E03.Lamentis.1080p.DSNP.WEB-DL.DDP5.1.H.264-TOMMY.chs&eng.srt")}, wantNil: false, wantErr: false, wantLang: language.ChineseSimpleEnglish, wantDialogueLen: 484},
		{name: "3", args: args{filePath: filepath.Join(testRootDir, "Bridge of Spies (2015) (1080p BluRay x265 Silence).zh-cn.srt")}, wantNil: false, wantErr: false, wantLang: language.ChineseSimpleEnglish, wantDialogueLen: 2069},
		{name: "4", args: args{filePath: filepath.Join(testRootDir, "Resident Evil Welcome to Raccoon City (2021) WEBRip-1080p.1.zh-cn.srt")}, wantNil: false, wantErr: false, wantLang: language.ChineseSimple, wantDialogueLen: 1472},
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

			fBytes, err := os.ReadFile(tt.args.filePath)
			if err != nil {
				t.Fatal(err)
			}
			inBytes, err := lan.ChangeFileCoding2UTF8(fBytes)
			if err != nil {

				t.Fatal(err)
			}
			dialogueCount := NewParser(log_helper.GetLogger4Tester()).parseContent(inBytes)
			if len(dialogueCount) != tt.wantDialogueLen || len(got.Dialogues) != tt.wantDialogueLen {
				t.Fatal("parse content dialogue error")
			}

			println(got.Name, got.Ext, got.Lang)
		})
	}
}
