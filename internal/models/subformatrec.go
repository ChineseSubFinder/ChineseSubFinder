package models

import "gorm.io/gorm"

// SubFormatRec 记录是否经过格式化，理论上只有一条
type SubFormatRec struct {
	gorm.Model
	FormatName int // 字幕格式化格式的名称（Normal or Emby 的枚举类型）
	Done       bool
}
