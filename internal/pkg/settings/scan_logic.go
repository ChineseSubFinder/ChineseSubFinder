package settings

type ScanLogic struct {
	SkipChineseMovie  bool `json:"skip_chinese_movie" default:"false"`  // 跳过中文的电影
	SkipChineseSeries bool `json:"skip_chinese_series" default:"false"` // 跳过中文的连续剧
}

func NewScanLogic(skipChineseMovie bool, skipChineseSeries bool) *ScanLogic {
	return &ScanLogic{SkipChineseMovie: skipChineseMovie, SkipChineseSeries: skipChineseSeries}
}
