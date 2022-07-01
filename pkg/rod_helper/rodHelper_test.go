package rod_helper

import (
	"testing"

	"github.com/allanpk716/ChineseSubFinder/pkg/log_helper"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

func TestNewBrowser(t *testing.T) {

	bPath, bok := launcher.LookPath()
	println(bok, bPath)
	desURL := "https://www.wikipedia.org/"
	httpProxyURL := ""
	browser, err := NewBrowser(log_helper.GetLogger4Tester(), "", httpProxyURL, true)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = browser.Close()
	}()
	page, err := browser.Page(proto.TargetCreateTarget{URL: desURL})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = page.Close()
	}()
}

func TestNewBrowserFromDocker(t *testing.T) {
	//desURL := "https://www.wikipedia.org/"
	//httpProxyURL := "http://127.0.0.1:10809"
	//remoteDockerURL := "ws://192.168.50.135:9222"
	//
	//browser, err := NewBrowserFromDocker(httpProxyURL, remoteDockerURL)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//defer func() {
	//	_ = browser.Close()
	//}()
	//page, err := browser.Page(proto.TargetCreateTarget{URL: desURL})
	//if err != nil {
	//	t.Fatal(err)
	//}
	//defer func() {
	//	_ = page.Close()
	//}()
}

func TestClear(t *testing.T) {
	Clear(log_helper.GetLogger4Tester())
}
