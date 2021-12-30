package sub_timeline_fiexer

type SubTimelineFixerConfig struct {
	// V1 的设置
	V1_MaxCompareDialogue int     // 最大需要匹配的连续对白，默认3
	V1_MaxStartTimeDiffSD float64 // 对白开始时间的统计 SD 最大误差，超过则不进行修正
	V1_MinMatchedPercent  float64 // 两个文件的匹配百分比（src/base），高于这个才比例进行修正
	V1_MinOffset          float64 // 超过这个(+-)偏移的时间轴才校正，否则跳过，单位秒
	// V2 的设置
	V2_SubOneUnitProcessTimeOut int     // 字幕时间轴校正一个单元的超时时间，单位秒
	V2_FrontAndEndPerBase       float64 // 前百分之 15 和后百分之 15 都不进行识别
	V2_FrontAndEndPerSrc        float64 // 前百分之 20 和后百分之 20 都不进行识别
	V2_WindowMatchPer           float64 // SrcSub 滑动窗体的占比
	V2_CompareParts             int     // 滑动窗体分段次数
	V2_FixThreads               int     // 字幕校正的并发线程
	V2_MaxStartTimeDiffSD       float64 // 对白开始时间的统计 SD 最大误差，超过则不进行修正
	V2_MinOffset                float64 // 超过这个(+-)偏移的时间轴才校正，否则跳过，单位秒
	V2_MaxOffsetTime            int     // 最大可以校正的时间偏移，时间是秒
}

// CheckDefault 检测默认值（比如某些之默认不能为0），不对就重置到默认值上
func (s *SubTimelineFixerConfig) CheckDefault() {
	// V1
	if s.V1_MaxCompareDialogue <= 0 {
		s.V1_MaxCompareDialogue = 3
	}
	if s.V1_MaxStartTimeDiffSD <= 0 {
		s.V1_MaxStartTimeDiffSD = 0.1
	}
	if s.V1_MinMatchedPercent <= 0 {
		s.V1_MinMatchedPercent = 0.1
	}
	if s.V1_MinOffset <= 0 {
		s.V1_MinOffset = 0.1
	}
	// V2
	if s.V2_SubOneUnitProcessTimeOut <= 0 {
		s.V2_SubOneUnitProcessTimeOut = 30
	}
	if s.V2_FrontAndEndPerBase <= 0 || s.V2_FrontAndEndPerBase >= 1.0 {
		s.V2_FrontAndEndPerBase = 0.15
	}
	if s.V2_FrontAndEndPerSrc <= 0 || s.V2_FrontAndEndPerSrc >= 1.0 {
		s.V2_FrontAndEndPerSrc = 0.2
	}
	if s.V2_WindowMatchPer <= 0 || s.V2_WindowMatchPer >= 1.0 {
		s.V2_WindowMatchPer = 0.7
	}
	if s.V2_CompareParts <= 0 {
		s.V2_CompareParts = 5
	}
	if s.V2_FixThreads <= 0 {
		s.V2_FixThreads = 3
	}
	if s.V2_MaxStartTimeDiffSD <= 0 {
		s.V2_MaxStartTimeDiffSD = 0.1
	}
	if s.V2_MinOffset <= 0 {
		s.V2_MinOffset = 0.1
	}

	if s.V2_MaxOffsetTime <= 0 {
		s.V2_MaxOffsetTime = 120
	}
}
