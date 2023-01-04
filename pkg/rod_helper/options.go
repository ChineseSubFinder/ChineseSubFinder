package rod_helper

import (
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
	"github.com/sirupsen/logrus"
)

type BrowserOptions struct {
	Log                  *logrus.Logger     // 日志
	LoadAdblock          bool               // 是否加载 adblock
	Settings             *settings.Settings // 配置
	preLoadUrl           string             // 预加载的url
	xrayPoolUrl          string             // xray pool url
	xrayPoolPort         string             // xray pool port
	browserInstanceCount int                // 浏览器最大的实例，xrayPoolUrl 有值的时候生效，用于爬虫。因为每启动一个实例就试用一个固定的代理，所以需要多个才行
}

func NewBrowserOptions(log *logrus.Logger, loadAdblock bool, settings *settings.Settings) *BrowserOptions {
	return &BrowserOptions{Log: log, LoadAdblock: loadAdblock, Settings: settings, browserInstanceCount: 1}
}

func (r *BrowserOptions) SetPreLoadUrl(url string) {
	r.preLoadUrl = url
}
func (r *BrowserOptions) PreLoadUrl() string {
	return r.preLoadUrl
}

// SetXrayPoolUrl 127.0.0.1
func (r *BrowserOptions) SetXrayPoolUrl(xrayUrl string) {
	r.xrayPoolUrl = xrayUrl
}

// XrayPoolUrl 127.0.0.1
func (r *BrowserOptions) XrayPoolUrl() string {
	return r.xrayPoolUrl
}

// SetXrayPoolPort 19035
func (r *BrowserOptions) SetXrayPoolPort(xrayPort string) {
	r.xrayPoolPort = xrayPort
}

// XrayPoolPort 19035
func (r *BrowserOptions) XrayPoolPort() string {
	return r.xrayPoolPort
}

func (r *BrowserOptions) SetBrowserInstanceCount(count int) {
	r.browserInstanceCount = count
}
func (r *BrowserOptions) BrowserInstanceCount() int {
	return r.browserInstanceCount
}
