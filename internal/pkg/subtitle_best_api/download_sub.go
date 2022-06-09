package subtitle_best_api

type DownloadSubReq struct {
	SubSha256     string `form:"sub_sha256"`       // 文件的 SHA256
	DownloadToken string `form:"download_token"`   // 下载令牌
	ApiKey        string `form:"api_key,optional"` // API Key，非必须，可能是某些用户才有的权限
}

type DownloadSubReply struct {
	Status     int    `json:"status"`                // 0 失败，1 成功。应该说正常就是下载文件了，失败才会使用这个结构体
	Message    string `json:"message"`               // 返回的信息，包括成功和失败的原因
	StoreRPath string `json:"store_r_path,optional"` // 存储路径
}
