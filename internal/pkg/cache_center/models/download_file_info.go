package models

import (
	"gorm.io/gorm"
	"time"
)

type DownloadFileInfo struct {
	gorm.Model
	UID            string    `gorm:"column:uid;primary_key"`
	RelPath        string    `gorm:"column:rel_path"`
	ExpirationTime time.Time `gorm:"column:expiration_time"`
}
