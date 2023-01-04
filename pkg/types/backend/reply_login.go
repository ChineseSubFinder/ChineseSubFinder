package backend

import "github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"

type ReplyLogin struct {
	AccessToken string            `json:"access_token,omitempty"` // 登录成功后返回令牌
	Settings    settings.Settings `json:"settings,omitempty"`     // 登录成功后返回当前的 Setting 信息
	Message     string            `json:"message,omitempty"`
}
