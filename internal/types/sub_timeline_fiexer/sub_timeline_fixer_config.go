package sub_timeline_fiexer

type SubTimelineFixerConfig struct {
	MaxCompareDialogue int     // 最大需要匹配的连续对白，默认3
	MaxStartTimeDiffSD float64 // 对白开始时间的统计 SD 最大误差，超过则不进行修正
	MinMatchedPercent  float64 // 两个文件的匹配百分比（src/base），高于这个才比例进行修正
	MinOffset          float64 // 超过这个(+-)偏移的时间轴才校正，否则跳过，单位秒
}

// CheckDefault 检测默认值（比如某些之默认不能为0），不对就重置到默认值上
func (s *SubTimelineFixerConfig) CheckDefault() {
	if s.MaxCompareDialogue == 0 {
		s.MaxCompareDialogue = 3
	}
	if s.MaxStartTimeDiffSD == 0 {
		s.MaxStartTimeDiffSD = 0.1
	}
	if s.MinMatchedPercent == 0 {
		s.MinMatchedPercent = 0.1
	}
	if s.MinOffset == 0 {
		s.MinOffset = 0.1
	}
}
