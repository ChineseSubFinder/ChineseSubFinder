package settings

type TmdbApiSettings struct {
	Enable bool   `json:"enable"`
	ApiKey string `json:"api_key"`
}
