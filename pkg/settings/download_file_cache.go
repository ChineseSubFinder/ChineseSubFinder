package settings

type DownloadFileCache struct {
	TTL  int    `json:"ttl" default:"4320"`  // 单位需要根据下面的单位转换，默认是小时的单位，然后是半年的过期时间
	Unit string `json:"unit" default:"hour"` // second, hour, 目前仅仅支持 秒和小时
}

func NewDownloadFileCache() *DownloadFileCache {
	return &DownloadFileCache{TTL: 4320, Unit: "hour"}
}

func (d DownloadFileCache) Check() {
	if d.Unit == "second" {
		// 半年的秒数
		if d.TTL < 259200 || d.TTL > 525600 {
			d.TTL = 259200
		}
	}
	if d.Unit == "hour" {
		// 半年的小时数
		if d.TTL < 4320 || d.TTL > 8760 {
			d.TTL = 4320
		}
	}
}
