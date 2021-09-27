package srt

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	"path"
	"testing"
)

func TestParser_DetermineFileType(t *testing.T) {
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
		{name: "1", args: args{filePath: path.Join(testRootDir, "[zimuku]_5_Loki.S01E02.The.Variant.1080p.DSNP.WEB-DL.DDP5.1.Atmos.H.264-CM.chs&eng.srt")}, wantNil: false, wantErr: false, wantLang: types.ChineseSimpleEnglish},
		{name: "2", args: args{filePath: path.Join(testRootDir, "[zimuku]_5_Loki.S01E03.Lamentis.1080p.DSNP.WEB-DL.DDP5.1.H.264-TOMMY.chs&eng.srt")}, wantNil: false, wantErr: false, wantLang: types.ChineseSimpleEnglish},
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

			} else {
				t.Fatal("DetermineFileTypeFromFile got:", got, "wantNil:", tt.wantNil)
			}

			println(got.Name, got.Ext, got.Lang)
		})
	}
}
