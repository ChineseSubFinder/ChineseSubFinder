package backend

type ReqCheckEmbyPath struct {
	AddressUrl    string `json:"address_url" binding:"required"`
	APIKey        string `json:"api_key" binding:"required"`
	PathType      string `json:"path_type" binding:"required"`
	CFSMediaPath  string `json:"cfs_media_path" binding:"required"`
	EmbyMediaPath string `json:"emby_media_path" binding:"required"`
}

type ReqCheckEmbyAPI struct {
	AddressUrl string `json:"address_url" binding:"required"`
	APIKey     string `json:"api_key" binding:"required"`
}
