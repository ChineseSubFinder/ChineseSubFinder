package sub_timeline_fiexer

type SubTimelineFixerConfig struct {
	MaxCompareDialogue int     // 最大需要匹配的连续对白，默认5
	MaxStartTimeDiffSD float64 // 对白开始时间的统计 SD 最大误差，超过则不进行修正
	MinMatchedPercent  float64 // 两个文件的匹配百分比（src/base），低于这个部进行修正
	MinOffset          float64 // 超过这个(+-)偏移的时间轴才校正，否则跳过，单位秒
}
