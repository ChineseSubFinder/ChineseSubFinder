package backend

type ReplyMovieList struct {
	Movies []MovieInfo `json:"movies"`
}

type MovieInfo struct {
	Name                     string `json:"name"`
	DirRootUrl               string `json:"dir_root_url"`
	VideoFPath               string `json:"video_f_path"`
	VideoUrl                 string `json:"video_url"`
	MediaServerInsideVideoID string `json:"media_server_inside_video_id"`
}
