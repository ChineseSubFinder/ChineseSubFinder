package chs_cht_changer

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/change_file_encode"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/unit_test_helper"
	"path/filepath"
	"testing"
)

func TestProcess(t *testing.T) {

	rootDir := unit_test_helper.GetTestDataResourceRootPath([]string{"sub_chs_cht_changer"}, 4, true)

	type args struct {
		srcSubFileFPath        string
		desChineseLanguageType int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "0", args: args{
			srcSubFileFPath:        filepath.Join(rootDir, "神枪手 (2021) 1080p DTSHD-MA.chinese(繁英,subhd).srt"),
			desChineseLanguageType: 0,
		}, wantErr: false},
		{name: "1", args: args{
			srcSubFileFPath:        filepath.Join(rootDir, "机动战士Z高达：星之继承者 (2005) 1080p TrueHD.chinese(繁).ssa"),
			desChineseLanguageType: 0,
		}, wantErr: false},
		{name: "2", args: args{
			srcSubFileFPath:        filepath.Join(rootDir, "机动战士Z高达Ⅲ：星辰的鼓动是爱 (2006) 1080p TrueHD.chinese(繁).ass"),
			desChineseLanguageType: 0,
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := change_file_encode.Process(tt.args.srcSubFileFPath, 0)
			if err != nil {
				t.Errorf("change_file_encode.Process() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err = Process(tt.args.srcSubFileFPath, tt.args.desChineseLanguageType); (err != nil) != tt.wantErr {
				t.Errorf("Process() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
