package models

import "gorm.io/gorm"

type UserInfo struct {
	gorm.Model
	Username string `json:"username" binding:"required,alphanum"`     // 用户名
	Password string `json:"password" binding:"required,min=6,max=12"` // 密码
}
