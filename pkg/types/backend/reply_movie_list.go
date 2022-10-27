package backend

type MovieInfo struct {
	Name                     string   `json:"name"`
	DirRootUrl               string   `json:"dir_root_url"`
	VideoFPath               string   `json:"video_f_path"`
	VideoUrl                 string   `json:"video_url"`
	MediaServerInsideVideoID string   `json:"media_server_inside_video_id"`
	SubFPathList             []string `json:"sub_f_path_list"`
	SubUrlList               []string `json:"sub_url_list"`
}

type MovieInfoV2 struct {
	Name             string `json:"name"`
	MainRootDirFPath string `json:"main_root_dir_f_path"` // x:\电影
	VideoFPath       string `json:"video_f_path"`         // x:\电影\壮志凌云\壮志凌云.mp4
}

type MovieSubsInfo struct {
	SubUrlList   []string `json:"sub_url_list"`
	SubFPathList []string `json:"sub_f_path_list"`
}
