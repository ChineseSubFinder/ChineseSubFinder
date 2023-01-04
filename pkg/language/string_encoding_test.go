package language

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/unit_test_helper"
)

func TestChangeFileCoding2UTF8(t *testing.T) {
	type args struct {
		subFileFPath string
	}
	tests := []struct {
		name                string
		args                args
		wantDesSubFileFPath string
		wantErr             bool
	}{
		{
			name: "00",
			args: args{
				subFileFPath: filepath.Join(unit_test_helper.GetTestDataResourceRootPath([]string{"change_sub_encode", "org-utf-8"}, 4, true),
					"Seven.Worlds.One.Planet.S01E01.Antarctica.2160p.BluRay.REMUX.HEVC.DTS-HD.MA.TrueHD.7.1.Atmos-FGT.chinese(简英,subhd).ass"),
			},
			wantDesSubFileFPath: filepath.Join(unit_test_helper.GetTestDataResourceRootPath([]string{"change_sub_encode", "org-utf-8"}, 4, true),
				"Seven.Worlds.One.Planet.S01E01.Antarctica.2160p.BluRay.REMUX.HEVC.DTS-HD.MA.TrueHD.7.1.Atmos-FGT.chinese(简英,subhd-utf-8).ass"),
			wantErr: false,
		},
		{
			name: "01",
			args: args{
				subFileFPath: filepath.Join(unit_test_helper.GetTestDataResourceRootPath([]string{"change_sub_encode", "org-utf-8"}, 4, true),
					"Seven.Worlds.One.Planet.S01E07.Africa.2160p.BluRay.REMUX.HEVC.DTS-HD.MA.TrueHD.7.1.Atmos-FGT.chinese(简英,zimuku).default.srt"),
			},
			wantDesSubFileFPath: filepath.Join(unit_test_helper.GetTestDataResourceRootPath([]string{"change_sub_encode", "org-utf-8"}, 4, true),
				"Seven.Worlds.One.Planet.S01E07.Africa.2160p.BluRay.REMUX.HEVC.DTS-HD.MA.TrueHD.7.1.Atmos-FGT.chinese(简英,zimuku-utf-8).default.srt"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			fBytes, err := os.ReadFile(tt.args.subFileFPath)
			if err != nil {
				t.Fatal(err)
			}
			got, err := ChangeFileCoding2UTF8(fBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("ChangeFileCoding2UTF8() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			wantDesFBytes, err := os.ReadFile(tt.wantDesSubFileFPath)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, wantDesFBytes) {
				t.Errorf("ChangeFileCoding2UTF8() got = %v, want %v", got, wantDesFBytes)
			}
		})
	}
}

func TestChangeFileCoding2GBK(t *testing.T) {
	type args struct {
		subFileFPath string
	}
	tests := []struct {
		name                string
		args                args
		wantDesSubFileFPath string
		wantErr             bool
	}{
		{
			name: "00",
			args: args{
				subFileFPath: filepath.Join(unit_test_helper.GetTestDataResourceRootPath([]string{"change_sub_encode", "org-utf-8"}, 4, true),
					"Seven.Worlds.One.Planet.S01E01.Antarctica.2160p.BluRay.REMUX.HEVC.DTS-HD.MA.TrueHD.7.1.Atmos-FGT.chinese(简英,subhd).ass"),
			},
			wantDesSubFileFPath: filepath.Join(unit_test_helper.GetTestDataResourceRootPath([]string{"change_sub_encode", "org-utf-8"}, 4, true),
				"Seven.Worlds.One.Planet.S01E01.Antarctica.2160p.BluRay.REMUX.HEVC.DTS-HD.MA.TrueHD.7.1.Atmos-FGT.chinese(简英,subhd-gbk).ass"),
			wantErr: false,
		},
		{
			name: "01",
			args: args{
				subFileFPath: filepath.Join(unit_test_helper.GetTestDataResourceRootPath([]string{"change_sub_encode", "org-utf-8"}, 4, true),
					"Seven.Worlds.One.Planet.S01E07.Africa.2160p.BluRay.REMUX.HEVC.DTS-HD.MA.TrueHD.7.1.Atmos-FGT.chinese(简英,zimuku).default.srt"),
			},
			wantDesSubFileFPath: filepath.Join(unit_test_helper.GetTestDataResourceRootPath([]string{"change_sub_encode", "org-utf-8"}, 4, true),
				"Seven.Worlds.One.Planet.S01E07.Africa.2160p.BluRay.REMUX.HEVC.DTS-HD.MA.TrueHD.7.1.Atmos-FGT.chinese(简英,zimuku-gbk).default.srt"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			fBytes, err := os.ReadFile(tt.args.subFileFPath)
			if err != nil {
				t.Fatal(err)
			}
			got, err := ChangeFileCoding2GBK(fBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("ChangeFileCoding2GBK() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			wantDesFBytes, err := os.ReadFile(tt.wantDesSubFileFPath)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, wantDesFBytes) {
				t.Errorf("ChangeFileCoding2GBK() got = %v, want %v", got, wantDesFBytes)
			}
		})
	}
}
