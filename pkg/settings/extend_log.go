package settings

type ExtendLog struct {
	SysLog SysLog
}

type SysLog struct {
	Enable   bool   `json:"enable"`
	Network  string `json:"network"`  // 留空就是本地 udp
	Address  string `json:"address"`  // 留空就是本地 localhost:514
	Priority int    `json:"priority"` // Debug 0, Info 1
	Tag      string `json:"tag"`
}
