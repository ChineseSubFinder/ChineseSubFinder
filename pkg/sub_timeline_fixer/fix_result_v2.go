package sub_timeline_fixer

import (
	"sync"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_helper"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/grd/stat"
)

type FixResult struct {
	StartVADIndex    int
	EndVADIndex      int
	OldMean          float64
	OldSD            float64
	NewMean          float64
	NewSD            float64
	Per              float64           // 占比
	IsOverParts      bool              // 是否有越接处
	MatchWindowInfos []MatchWindowInfo // 需要从 MatchInfo 的 IndexMatchWindowInfoMap 中按顺序提取
}

func (f FixResult) InRange(baseTimeDouble, timeStartDouble float64) (bool, float64) {

	startVad2Second := float64(f.StartVADIndex) / 100.0
	// 如果有越接处，因为是滑动窗体的原因，所以这个里的结束 index 并不是 FixResult 的，应该是具体的一个 MatchWindowInfo 的 EndIndex
	endVad2Second := float64(f.EndVADIndex) / 100.0
	if f.IsOverParts == true {
		endVad2Second = float64(f.MatchWindowInfos[0].EndVADIndex) / 100.0
	}
	if baseTimeDouble+startVad2Second <= timeStartDouble &&
		timeStartDouble <= baseTimeDouble+endVad2Second {
		// 在当前的范围内
		if f.IsOverParts == true {
			// 这里需要特殊处理，因为这个越接处，还需要二分
			for i := 0; i < len(f.MatchWindowInfos); i++ {
				b, newMean := f.chooseWhichWindow2Process(i, baseTimeDouble, timeStartDouble, startVad2Second)
				if b == true {
					return b, newMean
				}
			}

			return true, f.NewMean

		} else {
			// 无需特殊处理
			return true, f.NewMean

		}
	} else if timeStartDouble < baseTimeDouble+startVad2Second {
		// 小于当前的范围
		return true, f.NewMean
	} else {
		// 大于当前的范围
		return false, 0
	}
}

func (f FixResult) chooseWhichWindow2Process(index int, baseTimeDouble float64, timeStartDouble float64, startVad2Second float64) (bool, float64) {

	if f.MatchWindowInfos[index].OP.XLen <= 0 ||
		f.MatchWindowInfos[index].OP.YLen <= 0 {

		return false, 0
	}

	if timeStartDouble <= baseTimeDouble+startVad2Second+f.MatchWindowInfos[index].OP.XLen/100 {
		return true, f.MatchWindowInfos[index].OP.XMean
	} else {
		return true, f.MatchWindowInfos[index].OP.YMean
	}
}

// MatchInfo 匹配的信息
type MatchInfo struct {
	IndexMatchWindowInfoMap map[int]MatchWindowInfo // 匹配列表的顺序列表
	StartDiffTimeList       []float64
	StartDiffTimeMap        *treemap.Map
	StartDiffTimeListEx     stat.Float64Slice
}

type MatchWindowInfo struct {
	TimeDiffStartCorrelation float64 // 对白开始的时间偏移
	StartVADIndex            int
	EndVADIndex              int
	OP                       OverParts // 越接处信息
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
	Index            int                // 为了让并发处理的数据能够按顺序重新排序
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
	XLen  float64 // 分段处长度
	YLen  float64 // 分段处长度
	XMean float64 // X 段的 Mean 值
	YMean float64 // Y 段的 Mean 值
}
