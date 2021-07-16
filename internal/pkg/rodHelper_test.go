package pkg

import (
	"github.com/go-rod/rod/lib/proto"
	"testing"
	"time"
)

func TestLoadPage(t *testing.T) {
	desURL := "https://www.wikipedia.org/"
	httpProxyURL := "http://127.0.0.1:10809"
	_, err := NewBrowserLoadPage(desURL, httpProxyURL, 10*time.Second, 5)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoadPageFromRemoteDocker(t *testing.T) {
	desURL := "https://www.wikipedia.org/"
	httpProxyURL := "http://127.0.0.1:10809"
	remoteDockerURL := "ws://192.168.50.135:9222"
	_, err := NewBrowserLoadPageFromRemoteDocker(desURL, httpProxyURL, remoteDockerURL, 10*time.Second, 5)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoadPageByHijackRequests(t *testing.T) {
	desURL := "https://www.wikipedia.org/"
	httpProxyURL := "http://127.0.0.1:10809"
	_, err := NewBrowserLoadPageByHijackRequests(desURL, httpProxyURL, 10*time.Second, 5)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewBrowser(t *testing.T) {
	desURL := "https://www.wikipedia.org/"
	httpProxyURL := "http://127.0.0.1:10809"
	browser, err := NewBrowser(httpProxyURL)
	if err != nil {
		t.Fatal(err)
	}
	_, err = browser.Page(proto.TargetCreateTarget{URL: desURL})
	if err != nil {
		t.Fatal(err)
	}
	//err = rod.Try(func() {
	//	page.MustElement("#searchInput").MustInput("earth")
	//	page.MustElement("#search-form > fieldset > button").MustClick()
	//
	//	el := page.MustElement("#mw-content-text > div.mw-parser-output > table.infobox > tbody > tr:nth-child(1) > td > a > img")
	//	err = utils.OutputFile("b.png", el.MustResource())
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//})
	//if err != nil {
	//	t.Fatal(err)
	//}
}

func TestNewBrowserFromDocker(t *testing.T) {
	desURL := "https://www.wikipedia.org/"
	httpProxyURL := "http://127.0.0.1:10809"
	remoteDockerURL := "ws://192.168.50.135:9222"

	browser, err := NewBrowserFromDocker(httpProxyURL, remoteDockerURL)
	if err != nil {
		t.Fatal(err)
	}
	_, err = browser.Page(proto.TargetCreateTarget{URL: desURL})
	if err != nil {
		t.Fatal(err)
	}
}