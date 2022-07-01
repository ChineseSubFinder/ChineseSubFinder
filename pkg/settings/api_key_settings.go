package settings

type ApiKeySettings struct {
	Enabled bool   `json:"enabled"`
	Key     string `json:"key"`
}

func NewApiKeySettings(enabled bool, key string) *ApiKeySettings {
	return &ApiKeySettings{Enabled: enabled, Key: key}
}
