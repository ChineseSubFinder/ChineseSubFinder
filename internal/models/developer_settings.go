package models

import "gorm.io/gorm"

type DeveloperSettings struct {
	gorm.Model
	BarkServerUrl string // Bark 服务器的地址
}
