package models

import "gorm.io/gorm"

// DailyDownloadInfo 今日下载统计，每个下载源使用了多少的次数
type DailyDownloadInfo struct {
	gorm.Model
	SupplierName string `gorm:"column:supplier_name;type:varchar(255);not null" json:"supplier_name"`
	PublicIP     string `gorm:"column:public_ip;type:varchar(255);not null" json:"public_ip"`
	WhichDay     string `gorm:"column:which_day;type:varchar(255);not null" json:"which_day"`
	Count        int    `gor:"column:count;type:int;not null;default:0" json:"count"`
}
