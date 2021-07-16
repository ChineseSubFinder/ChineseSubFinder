package pkg

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"net/http"
	"net/url"
	"time"
)

/**
 * @Description: 			新建一个支持代理的 browser 对象
 * @param httpProxyURL		http://127.0.0.1:10809
 * @return *rod.Browser
 * @return error
 */
func NewBrowser(httpProxyURL string) (*rod.Browser, error) {
	var browser *rod.Browser
	err := rod.Try(func() {
		u := launcher.New().
			Proxy(httpProxyURL).
			MustLaunch()

		browser = rod.New().ControlURL(u).MustConnect()
	})
	if err != nil {
		return nil, err
	}

	return browser, nil
}

/**
 * @Description: 			访问目标 Url，返回 page，只是这个 page 有效，如果再次出发其他的事件无效
 * @param desURL			目标 Url
 * @param httpProxyURL		http://127.0.0.1:10809
 * @param timeOut			超时时间
 * @param maxRetryTimes		当是非超时 err 的时候，最多可以重试几次
 * @return *rod.Page
 * @return error
 */
func NewBrowserFromDocker(httpProxyURL, remoteDockerURL string) (*rod.Browser, error) {
	var browser *rod.Browser

	err := rod.Try(func() {
		l := launcher.MustNewRemote(remoteDockerURL)
		u := l.Proxy(httpProxyURL).MustLaunch()
		l.Headless(false).XVFB()
		browser = rod.New().Client(l.Client()).ControlURL(u).MustConnect()
	})
	if err != nil {
		return nil, err
	}

	return browser, nil
}

func NewPage(browser *rod.Browser) (*rod.Page, error) {
	page, err := browser.Page(proto.TargetCreateTarget{URL: ""})
	if err != nil {
		return nil, err
	}
	return page, err
}

func NewPageNavigate(browser *rod.Browser, desURL string, timeOut time.Duration, maxRetryTimes int) (*rod.Page, error) {

	page, err := NewPage(browser)
	if err != nil {
		return nil, err
	}
	page = page.Timeout(timeOut)
	nowRetryTimes := 0
	for nowRetryTimes <= maxRetryTimes {
		err = rod.Try(func() {
			wait := page.MustWaitNavigation()
			page.MustNavigate(desURL)
			wait()
		})
		if errors.Is(err, context.DeadlineExceeded) {
			// 超时
			return nil, err
		} else if err == nil {
			// 没有问题
			return page, nil
		}
	}
	return nil, err
}

func PageNavigate(page *rod.Page, desURL string, timeOut time.Duration, maxRetryTimes int) (*rod.Page, error) {
	var err error
	page = page.Timeout(timeOut)
	nowRetryTimes := 0
	for nowRetryTimes <= maxRetryTimes {
		err = rod.Try(func() {
			wait := page.MustWaitNavigation()
			page.MustNavigate(desURL)
			wait()
		})
		if errors.Is(err, context.DeadlineExceeded) {
			// 超时
			return nil, err
		} else if err == nil {
			// 没有问题
			return page, nil
		}
	}
	return nil, err
}

/**
 * @Description: 			访问目标 Url，返回 page，只是这个 page 有效，如果再次出发其他的事件无效
 * @param desURL			目标 Url
 * @param httpProxyURL		http://127.0.0.1:10809
 * @param timeOut			超时时间
 * @param maxRetryTimes		当是非超时 err 的时候，最多可以重试几次
 * @return *rod.Page
 * @return error
 */
func NewBrowserLoadPage(desURL string, httpProxyURL string, timeOut time.Duration, maxRetryTimes int) (*rod.Page, error) {
	browser, err := NewBrowser(httpProxyURL)
	if err != nil {
		return nil, err
	}
	page, err := browser.Page(proto.TargetCreateTarget{URL: ""})
	if err != nil {
		return nil, err
	}
	page = page.Timeout(timeOut)
	nowRetryTimes := 0
	for nowRetryTimes <= maxRetryTimes {
		err = rod.Try(func() {
			wait := page.MustWaitNavigation()
			page.MustNavigate(desURL)
			wait()
		})
		if errors.Is(err, context.DeadlineExceeded) {
			// 超时
			return nil, err
		} else if err == nil {
			// 没有问题
			return page, nil

		}
	}
	return nil, err
}

/**
 * @Description: 			访问目标 Url，返回 page，只是这个 page 有效，如果再次出发其他的事件无效
 * @param desURL			目标 Url
 * @param httpProxyURL		http://127.0.0.1:10809
 * @param timeOut			超时时间
 * @param maxRetryTimes		当是非超时 err 的时候，最多可以重试几次
 * @return *rod.Page
 * @return error
 */
func NewBrowserLoadPageFromRemoteDocker(desURL string, httpProxyURL, remoteDockerURL string, timeOut time.Duration, maxRetryTimes int) (*rod.Page, error) {
	browser, err := NewBrowserFromDocker(httpProxyURL, remoteDockerURL)
	if err != nil {
		return nil, err
	}
	page, err := browser.Page(proto.TargetCreateTarget{URL: ""})
	if err != nil {
		return nil, err
	}
	page = page.Timeout(timeOut)
	nowRetryTimes := 0
	for nowRetryTimes <= maxRetryTimes {
		err = rod.Try(func() {
			wait := page.MustWaitNavigation()
			page.MustNavigate(desURL)
			wait()
		})
		if errors.Is(err, context.DeadlineExceeded) {
			// 超时
			return nil, err
		} else if err == nil {
			// 没有问题
			break
		}
	}

	return page, nil
}

/**
 * @Description:			访问目标 Url，返回 page，只是这个 page 有效，如果再次出发其他的事件无效
 * @param desURL			目标 Url
 * @param httpProxyURL		http://127.0.0.1:10809
 * @param timeOut			超时时间
 * @param maxRetryTimes		当是非超时 err 的时候，最多可以重试几次
 * @return *rod.Page
 * @return error
 */
func NewBrowserLoadPageByHijackRequests(desURL string, httpProxyURL string, timeOut time.Duration, maxRetryTimes int) (*rod.Page, error) {

	var page *rod.Page
	var err error
	// 创建一个 page
	browser := rod.New()
	err = browser.Connect()
	if err != nil {
		return nil, err
	}
	page, err = browser.Page(proto.TargetCreateTarget{URL: ""})
	if err != nil {
		return nil, err
	}
	page = page.Timeout(timeOut)
	// 设置代理
	router := page.HijackRequests()
	defer router.Stop()

	err = rod.Try(func() {
		router.MustAdd("*", func(ctx *rod.Hijack) {
			px, _ := url.Parse(httpProxyURL)
			ctx.LoadResponse(&http.Client{
				Transport: &http.Transport{
					Proxy:           http.ProxyURL(px),
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
				Timeout: timeOut,
			}, true)
		})
	})
	if err != nil {
		return nil ,err
	}
	go router.Run()

	nowRetryTimes := 0
	for nowRetryTimes <= maxRetryTimes {
		err = rod.Try(func() {
			page.MustNavigate(desURL).MustWaitLoad()
		})
		if errors.Is(err, context.DeadlineExceeded) {
			// 超时
			return nil, err
		} else if err == nil {
			// 没有问题
			break
		}
		time.Sleep(time.Second)
		nowRetryTimes++
	}

	return page, nil
}
