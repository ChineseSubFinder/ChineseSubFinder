package settings

type TmdbApiSettings struct {
	Enable              bool   `json:"enable"`
	ApiKey              string `json:"api_key"`
	UseAlternateBaseURL bool   `json:"use_alternate_base_url"`
}

func NewTmdbApiSettings(enable bool, apiKey string, useAlternateBaseURL bool) *TmdbApiSettings {
	return &TmdbApiSettings{Enable: enable, ApiKey: apiKey, UseAlternateBaseURL: useAlternateBaseURL}
}
