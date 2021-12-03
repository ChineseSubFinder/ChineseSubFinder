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

func NewVADInfoBase(active bool, nowTime time.Duration) *VADInfo {
	return &VADInfo{
		Active: active,
		Time:   nowTime,
	}
}
