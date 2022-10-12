package backend

type SeasonInfo struct {
	Name          string         `json:"name"`
	RootDirPath   string         `json:"root_dir_path"`
	DirRootUrl    string         `json:"dir_root_url"`
	OneVideoInfos []OneVideoInfo `json:"one_video_info"`
}

type OneVideoInfo struct {
	Name                     string   `json:"name"`
	VideoFPath               string   `json:"video_f_path"`
	VideoUrl                 string   `json:"video_url"`
	Season                   int      `json:"season"`
	Episode                  int      `json:"episode"`
	SubFPathList             []string `json:"sub_f_path_list"`
	SubUrlList               []string `json:"sub_url_list"`
	MediaServerInsideVideoID string   `json:"media_server_inside_video_id"`
}

type SeasonInfoV2 struct {
	Name             string `json:"name"`
	MainRootDirFPath string `json:"main_root_dir_f_path"` // x:\连续剧
	RootDirPath      string `json:"root_dir_path"`        // x:\连续剧\绝命毒师
}
