package backend

type ReqCheckEmbyPath struct {
	AddressUrl    string `json:"address_url"  binding:"required"`
	APIKey        string `json:"api_key"  binding:"required"`
	CFSMediaPath  string `json:"cfs_media_path"  binding:"required"`
	EmbyMediaPath string `json:"emby_media_path"  binding:"required"`
}
