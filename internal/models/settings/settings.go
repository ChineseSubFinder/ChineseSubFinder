package settings

import "gorm.io/gorm"

type Settings struct {
	gorm.Model
	CommonSettings    CommonSettings    `gorm:"embedded;embeddedPrefix:common_"`
	AdvancedSettings  AdvancedSettings  `gorm:"embedded;embeddedPrefix:advanced_"`
	EmbySettings      EmbySettings      `gorm:"embedded;embeddedPrefix:emby_"`
	DeveloperSettings DeveloperSettings `gorm:"embedded;embeddedPrefix:developer_"`
}
