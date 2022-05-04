package backend

type ReplySeriesList struct {
	Movies []MovieInfo `json:"movies"`
}

type SeasonInfo struct {
	Name                     string `json:"name"`
	Url                      string `json:"dir_root_url"`
	MediaServerInsideVideoID string `json:"media_server_inside_video_id"`
}

type EpsInfo struct {
	Name                     string `json:"name"`
	Url                      string `json:"dir_root_url"`
	MediaServerInsideVideoID string `json:"media_server_inside_video_id"`
}
