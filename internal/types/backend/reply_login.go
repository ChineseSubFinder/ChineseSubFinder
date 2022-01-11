package backend

type ReplyLogin struct {
	AccessToken string `json:"access_token,omitempty"` // 登录成功后返回令牌
	Message     string `json:"message,omitempty"`
}
