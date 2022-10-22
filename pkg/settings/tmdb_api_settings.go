package settings

type TmdbApiSettings struct {
	Enable bool   `json:"enable"`
	ApiKey string `json:"api_key"`
}

func NewTmdbApiSettings(enable bool, apiKey string) *TmdbApiSettings {
	return &TmdbApiSettings{Enable: enable, ApiKey: apiKey}
}
