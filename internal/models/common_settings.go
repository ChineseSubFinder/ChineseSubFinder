package models

import "gorm.io/gorm"

type CommonSettings struct {
	gorm.Model
	UseHttpProxy     bool     // 是否使用 http 代理
	HttpProxyAddress string   // Http 代理地址，内网
	ScanInterval     string   // 一轮字幕扫描的间隔
	Threads          int      // 同时扫描的并发数
	RunScanAtStartUp bool     // 完成引导设置后，下次运行程序就开始扫描
	MoviePaths       []string // 电影的目录
	SeriesPaths      []string // 连续剧的目录
}
