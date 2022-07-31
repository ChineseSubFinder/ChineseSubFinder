package url_connectedness_helper

import (
	"crypto/tls"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// UrlConnectednessTest 测试输入 url 的连通性
func UrlConnectednessTest(testUrl, proxyAddr string) (bool, int64, error) {

	var httpClient http.Client
	if proxyAddr == "" {
		// 无需代理
		// 创建连接客户端
		httpClient = http.Client{
			Timeout: time.Second * testUrlTimeOut,
		}
	} else {
		// 需要代理
		// 检测代理iP访问地址
		//if proxyAddressValidHttpFormat(proxyAddr) == false {
		//	return false, 0, errors.New("proxy address illegal, only support http://xx:xx")
		//}
		// 解析代理地址
		proxy, err := url.Parse(proxyAddr)
		if err != nil {
			return false, 0, err
		}
		if strings.ToLower(proxy.Scheme) != "http" {
			return false, 0, errors.New("proxy address illegal, only support http://xx:xx")
		}
		// 设置网络传输
		netTransport := &http.Transport{
			Proxy:                 http.ProxyURL(proxy),
			MaxIdleConnsPerHost:   1000,
			ResponseHeaderTimeout: time.Second * time.Duration(testUrlTimeOut),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
		// 创建连接客户端
		httpClient = http.Client{
			Timeout:   time.Second * testUrlTimeOut,
			Transport: netTransport,
		}
	}

	begin := time.Now() //判断代理访问时间
	// 使用代理IP访问测试地址
	res, err := httpClient.Get(testUrl)
	if err != nil {
		return false, 0, err
	}
	defer func() {
		_ = res.Body.Close()
	}()
	speed := time.Now().Sub(begin).Nanoseconds() / 1000 / 1000 //ms
	// 判断是否成功访问，如果成功访问StatusCode应该为200
	if res.StatusCode != http.StatusOK {
		return false, 0, nil
	}
	return true, speed, nil
}

// proxyAddressValidHttpFormat 代理地址是否是有效的格式，必须是 http 的代理
func proxyAddressValidHttpFormat(proxyAddr string) bool {
	// 首先检测 proxyAddr 是否合法，必须是 http 的代理，不支持 https 代理
	re := regexp.MustCompile(`(http):\/\/[\w\-_]+(\.[\w\-_]+)+([\w\-\.,@?^=%&:/~\+#]*[\w\-\@?^=%&/~\+#])?`)
	result := re.FindAllStringSubmatch(proxyAddr, -1)
	if result == nil || len(result) < 1 {
		return false
	}
	return true
}

const testUrlTimeOut = 5

const (
	GoogleUrl = "https://google.com"
	BaiduUrl  = "https://baidu.com"
)
