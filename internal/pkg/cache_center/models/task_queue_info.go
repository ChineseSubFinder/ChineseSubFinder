package models

import "gorm.io/gorm"

type TaskQueueInfo struct {
	gorm.Model
	Priority int    `gorm:"column:priority"`
	RelPath  string `gorm:"column:rel_path"`
}
