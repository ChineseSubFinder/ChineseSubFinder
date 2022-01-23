package backend

type ReqCheckProxy struct {
	HttpProxyUrl string `json:"http_proxy_url"  binding:"required"`
}
