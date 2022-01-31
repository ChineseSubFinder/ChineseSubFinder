package settings

type TimelineFixerSettings struct {
	MaxOffsetTime int     `json:"max_offset_time"` // 最大支持校正时间偏移的范围，单位秒
	MinOffset     float64 `json:"min_offset"`      // 最小的时间片校正偏移，低于这个（正负）就跳过不校正，单位秒
}

func NewTimelineFixerSettings() *TimelineFixerSettings {
	return &TimelineFixerSettings{}
}

func (t *TimelineFixerSettings) Check() {
	if t.MaxOffsetTime <= 0 || t.MaxOffsetTime > 120 {
		t.MaxOffsetTime = 60 // 60s
	}

	if t.MinOffset <= 0 || t.MinOffset > 1 {
		t.MinOffset = 0.1 // 100ms
	}
}
