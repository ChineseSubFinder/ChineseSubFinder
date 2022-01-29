package settings

type DeveloperSettings struct {
	Enable            bool   `json:"enable"`              // 是否启用
	BarkServerAddress string `json:"bark_server_address"` // Bark 服务器的地址
}

func NewDeveloperSettings() *DeveloperSettings {
	return &DeveloperSettings{}
}
