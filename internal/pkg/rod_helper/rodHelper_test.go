package rod_helper

import (
	"github.com/go-rod/rod/lib/proto"
	"testing"
)

func TestNewBrowser(t *testing.T) {
	desURL := "https://www.wikipedia.org/"
	httpProxyURL := "http://127.0.0.1:10809"
	browser, err := NewBrowser(httpProxyURL, true)
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
	desURL := "https://www.wikipedia.org/"
	httpProxyURL := "http://127.0.0.1:10809"
	remoteDockerURL := "ws://192.168.50.135:9222"

	browser, err := NewBrowserFromDocker(httpProxyURL, remoteDockerURL)
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

func TestClear(t *testing.T) {
	Clear()
}
