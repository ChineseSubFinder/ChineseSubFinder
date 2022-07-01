package url_connectedness_helper

import "testing"

func TestUrlConnectednessTest(t *testing.T) {
	type args struct {
		testUrl   string
		proxyAddr string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{name: "0", args: args{
			testUrl:   "https://google.com",
			proxyAddr: "",
		}, want: false, wantErr: true},
		{name: "1", args: args{
			testUrl:   "https://google.com",
			proxyAddr: "",
		}, want: false, wantErr: true},
		{name: "2", args: args{
			testUrl:   "https://google.com",
			proxyAddr: "",
		}, want: false, wantErr: true},
		{name: "3", args: args{
			testUrl:   "https://google.com",
			proxyAddr: "",
		}, want: false, wantErr: true},
		{name: "4", args: args{
			testUrl:   "https://google.com",
			proxyAddr: "http://192.168.50.252:20172",
		}, want: true, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := UrlConnectednessTest(tt.args.testUrl, tt.args.proxyAddr)
			if (err != nil) != tt.wantErr {
				t.Errorf("UrlConnectednessTest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("UrlConnectednessTest() got = %v, want %v", got, tt.want)
			}
		})
	}
}
