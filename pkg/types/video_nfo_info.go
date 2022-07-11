package types

import "strconv"

// VideoNfoInfo 从 movie.xml *.nfo 中解析出的视频信息
type VideoNfoInfo struct {
	ImdbId        string
	TmdbId        string
	TVdbId        string
	Season        int
	Episode       int
	Year          string
	Title         string
	OriginalTitle string
	ReleaseDate   string
	IsMovie       bool
}

func (v *VideoNfoInfo) GetYear() int {

	atoi, err := strconv.Atoi(v.Year)
	if err != nil {
		return 0
	}
	return atoi
}
