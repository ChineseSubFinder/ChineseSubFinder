package models

import "gorm.io/gorm"

// SubFormatRec 记录是否经过格式化，理论上只有一条
type SubFormatRec struct {
	gorm.Model
	FormatName string // 字幕格式化的名称
	Done       bool
}
