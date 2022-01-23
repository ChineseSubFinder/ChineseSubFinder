package settings

type DeveloperSettings struct {
	BarkServerAddress string `json:"bark_server_address"` // Bark 服务器的地址
}

func NewDeveloperSettings() *DeveloperSettings {
	return &DeveloperSettings{}
}
