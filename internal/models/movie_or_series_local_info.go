package models

type MovieOrSeriesLocalInfo struct {
	IsMovie      bool   `json:"is_movie"`                         // 不是电影就是连续剧
	Season       int    `json:"season"`                           // 季度
	RootDirRPath string `json:"root_dir_r_path"`                  // 这个电影或者连续剧（不是季的文件夹，而是这个连续剧的目录）的相对路径
	IMDBInfoID   string `json:"imdb_info_id"  binding:"required"` // IMDB ID
}

func NewMovieOrSeriesLocalInfo(isMovie bool, season int, rootDirRPath string) *MovieOrSeriesLocalInfo {
	return &MovieOrSeriesLocalInfo{IsMovie: isMovie, Season: season, RootDirRPath: rootDirRPath}
}
