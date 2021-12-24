package sub_timeline_fixer

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_helper"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/grd/stat"
	"sync"
)

type FixResult struct {
	StartVADIndex int
	EndVADIndex   int
	OldMean       float64
	OldSD         float64
	NewMean       float64
	NewSD         float64
	Per           float64   // 占比
	OP            OverParts // 越接处信息
}

func (f FixResult) InRange(baseTimeDouble, timeStartDouble float64) (bool, float64) {

	startVad2Second := f.StartVADIndex / 100
	endVad2Second := f.EndVADIndex / 100

	if baseTimeDouble+float64(startVad2Second) <= timeStartDouble &&
		timeStartDouble <= baseTimeDouble+float64(endVad2Second) {
		// 在当前的范围内
		if f.OP.Has == true {
			// 这里需要特殊处理，因为这个越接处，还需要二分
			if timeStartDouble <= baseTimeDouble+float64(startVad2Second)+f.OP.XLen/100 {
				return true, f.OP.XMean
			} else {
				return true, f.OP.YMean
			}
		} else {
			// 无需特殊处理
			return true, f.NewMean

		}
	} else if timeStartDouble < baseTimeDouble+float64(startVad2Second) {
		// 小于当前的范围
		return true, f.NewMean
	} else {
		// 大于当前的范围
		return false, 0
	}
}

// MatchInfo 匹配的信息
type MatchInfo struct {
	StartDiffTimeList   []float64
	StartDiffTimeMap    *treemap.Map
	StartDiffTimeListEx stat.Float64Slice
}

// WindowInfo 滑动窗体信息
type WindowInfo struct {
	BaseAudioFloatList []float64           // 基准 VAD
	BaseUnit           *sub_helper.SubUnit // 基准 VAD
	SrcUnit            *sub_helper.SubUnit // 需要匹配的 VAD
	MatchedTimes       int                 // 匹配上的次数
	SrcWindowLen       int                 // 滑动窗体长度
	SrcSlideStartIndex int                 // 滑动起始索引
	SrcSlideLen        int                 // 滑动距离
	OneStep            int                 // 每次滑动的长度
}

// InputData 修复函数传入多线程的数据结构
type InputData struct {
	BaseUnit         sub_helper.SubUnit // 基准 VAD
	BaseAudioVADList []float64          // 基准 VAD
	SrcUnit          sub_helper.SubUnit // 需要匹配的 VAD
	OffsetIndex      int                // 滑动窗体的移动偏移索引
	Wg               *sync.WaitGroup    // 并发锁
}

// SubVADBlockInfo 字幕分块信息
type SubVADBlockInfo struct {
	Index      int
	StartIndex int
	EndIndex   int
}

/*
	OverParts 总长度 D = XLen + YLen
*/
type OverParts struct {
	Has   bool    // 是否有越接处
	XLen  float64 // 分段处长度
	YLen  float64 // 分段处长度
	XMean float64 // X 段的 Mean 值
	YMean float64 // Y 段的 Mean 值
}
