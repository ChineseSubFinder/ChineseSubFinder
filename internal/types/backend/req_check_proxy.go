package backend

import "github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"

type ReqCheckProxy struct {
	ProxySettings settings.ProxySettings `json:"proxy_settings"  binding:"required"`
}
