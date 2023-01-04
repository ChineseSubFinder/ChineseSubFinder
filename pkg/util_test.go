package pkg

import (
	"path/filepath"
	"testing"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/log_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/unit_test_helper"
)

func TestCloseChrome(t *testing.T) {

	// BUG: will produce Logs under this dir
	CloseChrome(log_helper.GetLogger4Tester())
}

func TestFileNameIsBDMV(t *testing.T) {

	rootDir := unit_test_helper.GetTestDataResourceRootPath([]string{"movies", "失控玩家 (2021)"}, 4, false)
	dbmvFPath := filepath.Join(rootDir, "CERTIFICATE", "id.bdmv")
	bok, fakeVideoFPath := FileNameIsBDMV(dbmvFPath)
	if bok == false {
		t.Fatal("FileNameIsBDMV error")
	}
	println(fakeVideoFPath)
}

func TestGetRestOfDaySec(t *testing.T) {

	rest := GetRestOfDaySec()
	println(rest)
}

func TestGetPublicIP(t *testing.T) {

	//got := GetPublicIP(log_helper.GetLogger4Tester(), settings.NewTaskQueue())
	//println("NoProxy:", got)
	//
	//sock5ProxySettings := settings.NewProxySettings(true, "socks5", local_http_proxy_server.LocalHttpProxyPort,
	//	"127.0.0.1", "10808", "", "")
	//
	//got = GetPublicIP(log_helper.GetLogger4Tester(), settings.NewTaskQueue(), sock5ProxySettings)
	//println("UseProxy socks5:", got)
	//err := sock5ProxySettings.CloseLocalHttpProxyServer()
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//httpProxySettings := settings.NewProxySettings(true, "http", local_http_proxy_server.LocalHttpProxyPort,
	//	"127.0.0.1", "10809", "", "")
	//got = GetPublicIP(log_helper.GetLogger4Tester(), settings.NewTaskQueue(), httpProxySettings)
	//println("UseProxy http:", got)
	//err = httpProxySettings.CloseLocalHttpProxyServer()
	//if err != nil {
	//	t.Fatal(err)
	//}
}

func TestSortByModTime(t *testing.T) {
	//type args struct {
	//	fileList []string
	//}
	//tests := []struct {
	//	name string
	//	args args
	//	want []string
	//}{
	//	{name: "001", args: args{fileList: []string{
	//		"X:\\电影\\21座桥 (2019)\\21座桥 (2019) 720p AAC.mp4",
	//		"X:\\电影\\Texas Chainsaw Massacre (2022)\\Texas Chainsaw Massacre (2022) WEBDL-1080p.mkv",
	//		"X:\\电影\\76 Days (2020)\\76 Days (2020) WEBDL-1080p.mkv"}},
	//		want: []string{
	//			"a",
	//			"b",
	//			"c"}},
	//}
	//for _, tt := range tests {
	//	t.Run(tt.name, func(t *testing.T) {
	//		if got := SortByModTime(tt.args.fileList); !reflect.DeepEqual(got, tt.want) {
	//			t.Errorf("SortByModTime() = %v, want %v", got, tt.want)
	//		}
	//	})
	//}
}
