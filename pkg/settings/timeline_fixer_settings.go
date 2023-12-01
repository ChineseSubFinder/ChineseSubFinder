package settings

type TimelineFixerSettings struct {
	MaxOffsetTime int     `json:"max_offset_time"` // 最大支持校正时间偏移的范围，单位秒
	MinOffset     float64 `json:"min_offset"`      // 最小的时间片校正偏移，低于这个（正负）就跳过不校正，单位秒
	ThreadCount   int     `json:"thread_count"`    // 线程数
}

func NewTimelineFixerSettings() *TimelineFixerSettings {
	return &TimelineFixerSettings{
		MaxOffsetTime: 700,
		MinOffset:     0.2,
		ThreadCount:   5,
	}
}

func (t *TimelineFixerSettings) Check() {
	if t.MaxOffsetTime <= 0 || t.MaxOffsetTime > 700 {
		t.MaxOffsetTime = 700 // 60s
	}

	if t.MinOffset <= 0 || t.MinOffset > 5 {
		t.MinOffset = 0.2 // 100ms
	}

	if t.ThreadCount <= 0 || t.ThreadCount > 8 {
		t.ThreadCount = 5
	}
}
