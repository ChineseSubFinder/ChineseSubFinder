package models

import "gorm.io/gorm"

type HotFix struct {
	gorm.Model
	Key  string // Hotfix Key 针对修复的具体问题
	Done bool
}
