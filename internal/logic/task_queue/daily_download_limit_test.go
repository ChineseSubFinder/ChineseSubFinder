package task_queue

import (
	"testing"
)

const supplieName = "testtest"

func TestGetDailyDownloadCount(t *testing.T) {
	type args struct {
		supplierName string
		whichDay     []string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{name: "00", args: args{
			supplierName: supplieName,
			whichDay:     nil,
		}, want: 0, wantErr: false},
		{name: "01", args: args{
			supplierName: supplieName,
			whichDay:     []string{"2022-04-01"},
		}, want: 0, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := GetDailyDownloadCount(tt.args.supplierName, tt.args.whichDay...)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDailyDownloadCount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetDailyDownloadCount() got = %v, want %v", got, tt.want)
			}

			//err = DelBucket(tt.args.supplierName)
			//if err != nil {
			//	t.Errorf("DelBucket(), err = %v", err.Error())
			//}
		})
	}
}

func TestAddDailyDownloadCount(t *testing.T) {
	type args struct {
		supplierName string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{name: "00", args: args{
			supplierName: supplieName,
		}, want: 1, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := AddDailyDownloadCount(tt.args.supplierName)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddDailyDownloadCount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AddDailyDownloadCount() got = %v, want %v", got, tt.want)
			}

		})
	}
}
