package backend

import "github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"

type ReqCheckProxy struct {
	ProxySettings settings.ProxySettings `json:"proxy_settings"  binding:"required"`
}
