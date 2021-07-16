package ass

import (
	"testing"
)

func TestParser_DetermineFileTypeFromFile(t *testing.T) {

	type args struct {
		filePath string
	}
	tests := []struct {
		name    string
		args    args
		wantNil bool
		wantErr bool
	}{
		{name: "1", args: args{filePath: "C:\\Tmp\\Loki - S01E01 - Glorious Purpose WEBDL-1080p Proper.chs[subhd].ass"}, wantNil: false, wantErr: false},
		{name: "3", args: args{filePath: "C:\\tmp\\oslo.2021.1080p.web.h264-naisu.简体&英文.ass"}, wantNil: false, wantErr: false},
		{name: "4", args: args{filePath: "C:\\Tmp\\oslo.2021.1080p.web.h264-naisu.繁体&英文.ass"}, wantNil: false, wantErr: false},
		{name: "5", args: args{filePath: "C:\\Tmp\\oslo.2021.1080p.web.h264-naisu.繁体.ass"}, wantNil: false, wantErr: false},
		{name: "6", args: args{filePath: "X:\\连续剧\\黑镜 (2011)\\Season 2\\黑镜 - S02E02 - 白熊.en.ass"}, wantNil: false, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Parser{}
			got, err := p.DetermineFileTypeFromFile(tt.args.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("DetermineFileTypeFromFile() error = %v, wantErr %v", err, tt.wantErr)
				t.Fatal(err)
				return
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