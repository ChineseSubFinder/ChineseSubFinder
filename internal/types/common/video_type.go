package common

type VideoType int

const (
	Movie  VideoType = iota // 电影
	Series                  // 连续剧，可能需要分美剧、日剧、韩剧？
	Anime                   // 动画
)

func (c VideoType) String() string {
	switch c {
	case Movie:
		return "movie"
	case Series:
		return "series"
	case Anime:
		return "anime"
	}
	return "N/A"
}
