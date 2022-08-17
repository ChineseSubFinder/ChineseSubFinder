package rod_helper

import (
	"testing"
	"time"

	"github.com/allanpk716/ChineseSubFinder/pkg/log_helper"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

func TestNewBrowser(t *testing.T) {

	bPath, bok := launcher.LookPath()
	println(bok, bPath)
	desURL := "https://google.com"
	httpProxyURL := "http://127.0.0.1:63204"
	browser, err := NewBrowserBase(log_helper.GetLogger4Tester(), "", httpProxyURL, true)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = browser.Close()
	}()

	page, status, responseUrl, err := NewPageNavigateWithProxy(browser, httpProxyURL, desURL, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}

	println("Status:", status, "ResponseUrl:", responseUrl)

	page, err = browser.Page(proto.TargetCreateTarget{URL: desURL})
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
	//browser, err := NewBrowserBaseFromDocker(httpProxyURL, remoteDockerURL)
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
