package types

// VideoIMDBInfo 从 movie.xml *.nfo 中解析出的视频信息
type VideoIMDBInfo struct {
	ImdbId        string
	TVdbId        string
	Year          string
	Title         string
	OriginalTitle string
	ReleaseDate   string
}
