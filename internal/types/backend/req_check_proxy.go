package backend

type ReqCheckProxy struct {
	HttpProxyAddress string `json:"http_proxy_address"  binding:"required"`
}
