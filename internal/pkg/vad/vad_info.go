package vad

import (
	"time"
)

type VADInfo struct {
	Frame  int           // 第几帧
	Offset int           // 音频的偏移
	Active bool          // 当前帧（时间窗口）是否检测到语音
	Time   time.Duration // 时间点
}

func NewVADInfo(frame, offset int, active bool, nowTime time.Duration) *VADInfo {
	return &VADInfo{
		Frame:  frame,
		Offset: offset,
		Active: active,
		Time:   nowTime,
	}
}

// GetTimeRange 获取这个 VAD 实例从 startTime，开始，向后多少 ms 的时间段的 VAD 新实例
func GetTimeRange(inVADInfos []VADInfo, starttime, timeRange int) []VADInfo {

	var outVADInfos = make([]VADInfo, 0)

	startTime := time.Duration(starttime)
	endTime := time.Duration(starttime + timeRange)

	for _, inVADInfo := range inVADInfos {

		if inVADInfo.Time < startTime || inVADInfo.Time > endTime {
			continue
		}
		outVADInfos = append(outVADInfos, inVADInfo)
	}

	return outVADInfos
}

// InsertVADInfo 得到的是 VAD 状态变换的节点，中间缺失了连续的 VAD 点信息，使用本函数可以进行插值
func InsertVADInfo(inVADInfos []VADInfo, duration int) []VADInfo {

	var outVADInfos = make([]VADInfo, 0)

	// 找到第一句，从这个 StartTime 之前标记为 VAD false
	if inVADInfos[0].Time != 0 {

	}

	return outVADInfos
}

const (
	// Mode vad mode，VAD 的模式
	Mode = 2
	// FrameDuration frame duration，分析的时间窗口
	FrameDuration = 10
)
