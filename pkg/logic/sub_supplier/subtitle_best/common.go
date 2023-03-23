package subtitle_best

type Subtitle struct {
	SubSha256 string `json:"sub_sha256"`
	Title     string `json:"title"`
	Language  int    `json:"language"`
	Ext       string `json:"ext"`
	IsMovie   bool   `json:"is_movie"`
	Season    int    `json:"season"`
	Episode   int    `json:"episode"`
	Token     string `json:"token"`
}

type SubtitleResponse struct {
	Status    int        `json:"status"`
	Message   string     `json:"message"`
	Subtitles []Subtitle `json:"subtitles"`
}

type SeasonPackagesResponse struct {
	Status           int      `json:"status"`
	Message          string   `json:"message"`
	SeasonPackageIds []string `json:"season_package_ids"`
}

type GetUrlResponse struct {
	Status       int    `json:"status"`
	Message      string `json:"message"`
	DownloadLink string `json:"download_link"`
}

type SearchMovieSubtitleRequest struct {
	ImdbID string `json:"imdb_id"`
	ApiKey string `json:"api_key"`
}

type SearchTVEpsSubtitleRequest struct {
	ImdbID  string `json:"imdb_id"`
	Season  int    `json:"season"`
	Episode int    `json:"episode"`
	ApiKey  string `json:"api_key"`
}

type SearchTVSeasonPackagesRequest struct {
	ImdbID string `json:"imdb_id"`
	Season int    `json:"season"`
	ApiKey string `json:"api_key"`
}

type SearchTVSeasonPackageByIDRequest struct {
	ImdbID          string `json:"imdb_id"`
	SeasonPackageId string `json:"season_package_id"`
	ApiKey          string `json:"api_key"`
}

type DownloadUrlConvertRequest struct {
	SubSha256       string `json:"sub_sha256"`
	ImdbID          string `json:"imdb_id"`
	IsMovie         bool   `json:"is_movie"`
	Season          int    `json:"season"`
	Episode         int    `json:"episode"`
	SeasonPackageId string `json:"season_package_id"`
	Language        int    `json:"language"`
	ApiKey          string `json:"api_key"`
	Token           string `json:"token"`
}
