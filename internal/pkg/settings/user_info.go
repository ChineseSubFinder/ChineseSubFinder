package settings

type UserInfo struct {
	Username string `json:"username" binding:"required,alphanum"`     // 用户名
	Password string `json:"password" binding:"required,min=6,max=30"` // 密码
}

func NewUserInfo(userName, password string) *UserInfo {
	return &UserInfo{
		Username: userName,
		Password: password,
	}
}
