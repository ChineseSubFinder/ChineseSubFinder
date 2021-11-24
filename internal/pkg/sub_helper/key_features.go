package sub_helper

// KeyFeatures 钥匙的组合特征
type KeyFeatures struct {
	Big    Feature // 大锯齿
	Middle Feature // 中锯齿
	Small  Feature // 小锯齿
}

func NewKeyFeatures(big, middle, small Feature) KeyFeatures {
	return KeyFeatures{
		big, middle, small,
	}
}

// Feature 钥匙锯齿的特征
type Feature struct {
	BigThanTime  float64 // 大于这个时间
	LessThanTime float64 // 小于这个时间
	LeastCount   int     // 至少有几个特征
	NowCount     int     // 用于计数
}

// NewFeature 时间如果传入的是 -1，那么就跳过这个判断
func NewFeature(BigThanTime, LessThanTime float64, LeastCount int) Feature {
	return Feature{
		BigThanTime, LessThanTime, LeastCount, 0,
	}
}

// Match 判断这个间隔是否符合要求
func (f Feature) Match(interval float64) bool {
	if interval > f.BigThanTime && interval < f.LessThanTime {
		return true
	} else {
		return false
	}
}
