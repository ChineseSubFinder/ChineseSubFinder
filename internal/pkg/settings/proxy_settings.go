package settings

type ProxySettings struct {
	UseSocks5OrHttpProxy bool   `json:"use_socks5_or_http_proxy"` // 是使用 socks5(true) 还是 http(false) 代理，默认是 http
	UseHttpProxy         bool   `json:"use_http_proxy"`           // 是否使用 http 代理
	HttpProxyAddress     string `json:"http_proxy_address"`       // Http 代理地址，内网
	UseSocks5Proxy       bool   `json:"use_socks5_proxy"`         // 是否使用 socks5 代理
	Socks5ProxyAddress   string `json:"socks5_proxy_address"`     // Socks5 代理地址，内网
	Referer              string `json:"-"`                        // 可能下载文件的时候需要设置
}

func NewProxySettings(useSocks5OrHttpProxy bool, useHttpProxy bool, httpProxyAddress string, useSocks5Proxy bool, socks5ProxyAddress string) *ProxySettings {
	return &ProxySettings{UseSocks5OrHttpProxy: useSocks5OrHttpProxy, UseHttpProxy: useHttpProxy, HttpProxyAddress: httpProxyAddress, UseSocks5Proxy: useSocks5Proxy, Socks5ProxyAddress: socks5ProxyAddress}
}
