package models

type TaskQueueInfo struct {
	Priority int    `gorm:"column:priority"`
	RelPath  string `gorm:"column:rel_path"`
}
