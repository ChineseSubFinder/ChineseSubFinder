package common

// VideoInfo 从 movie.xml *.nfo 中解析出的视频信息
type VideoInfo struct {
	ImdbId string
	TVdbId string
	Year string
	Title string
	OriginalTitle string
}
