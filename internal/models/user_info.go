package models

import "gorm.io/gorm"

type UserInfo struct {
	gorm.Model
	Username string `json:"username"` // 用户名
	Password string `json:"password"` // 密码
}
