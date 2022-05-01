package models

import "time"

type DownloadFileInfo struct {
	UID            string    `gorm:"column:uid;primary_key"`
	RelPath        string    `gorm:"column:rel_path"`
	ExpirationTime time.Time `gorm:"column:expiration_time"`
}
