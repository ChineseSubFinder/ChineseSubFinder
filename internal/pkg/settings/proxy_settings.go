package settings

type ProxySettings struct {
	UseHttpProxy     bool   `json:"use_http_proxy"`     // 是否使用 http 代理
	HttpProxyAddress string `json:"http_proxy_address"` // Http 代理地址，内网
	Referer          string `json:"-"`                  // 可能下载文件的时候需要设置
}
