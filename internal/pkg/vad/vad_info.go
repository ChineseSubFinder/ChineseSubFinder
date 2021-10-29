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

const (
	// Mode vad mode，VAD 的模式
	Mode = 2
	// FrameDuration frame duration，分析的时间窗口
	FrameDuration = 10
)
