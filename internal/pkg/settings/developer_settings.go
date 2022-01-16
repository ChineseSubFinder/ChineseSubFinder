package settings

type DeveloperSettings struct {
	BarkServerUrl string `json:"bark_server_url"` // Bark 服务器的地址
}

func NewDeveloperSettings() *DeveloperSettings {
	return &DeveloperSettings{}
}
