package path_helper

import "testing"

func Test_fixSMBPath(t *testing.T) {
	type args struct {
		orgPath string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "smb-00", args: args{
			orgPath: "smb://192.168.1.12/haha",
		}, want: "smb://192.168.1.12/haha"},
		{name: "smb-01", args: args{
			orgPath: "smb:/192.168.1.12/haha",
		}, want: "smb://192.168.1.12/haha"},
		{name: "smb-02", args: args{
			orgPath: "smb://192.168.1.12/haha\\test",
		}, want: "smb://192.168.1.12/haha\\test"},
		{name: "smb-03", args: args{
			orgPath: "smb:/192.168.1.12/haha\\test",
		}, want: "smb://192.168.1.12/haha\\test"},

		{name: "afp-00", args: args{
			orgPath: "afp://192.168.1.12/haha",
		}, want: "afp://192.168.1.12/haha"},
		{name: "afp-01", args: args{
			orgPath: "afp:/192.168.1.12/haha",
		}, want: "afp://192.168.1.12/haha"},
		{name: "afp-02", args: args{
			orgPath: "afp://192.168.1.12/haha\\test",
		}, want: "afp://192.168.1.12/haha\\test"},
		{name: "afp-03", args: args{
			orgPath: "afp:/192.168.1.12/haha\\test",
		}, want: "afp://192.168.1.12/haha\\test"},

		{name: "nfs-00", args: args{
			orgPath: "nfs://192.168.1.12/haha",
		}, want: "nfs://192.168.1.12/haha"},
		{name: "nfs-01", args: args{
			orgPath: "nfs:/192.168.1.12/haha",
		}, want: "nfs://192.168.1.12/haha"},
		{name: "nfs-02", args: args{
			orgPath: "nfs://192.168.1.12/haha\\test",
		}, want: "nfs://192.168.1.12/haha\\test"},
		{name: "nfs-03", args: args{
			orgPath: "nfs:/192.168.1.12/haha\\test",
		}, want: "nfs://192.168.1.12/haha\\test"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FixShareFileProtocolsPath(tt.args.orgPath); got != tt.want {
				t.Errorf("FixShareFileProtocolsPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
