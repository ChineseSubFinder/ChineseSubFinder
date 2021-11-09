package sub_helper

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/vad"
	"math"
	"time"
)

type SubUnit struct {
	baseTime        time.Time // 这个是基础的时间，后续需要减去这个，不然与导出的片段字幕去对比会有一个起始时间的偏差
	offsetStartTime time.Time // 相对时间，这个时间会减去 baseTime 再存储
	offsetEndTime   time.Time // 相对时间，这个时间会减去 baseTime 再存储
	VADList         []vad.VADInfo
	subCount        int
	firstAdd        bool
	outVADBytes     []byte
}

func NewSubUnit() *SubUnit {
	return &SubUnit{
		VADList:     make([]vad.VADInfo, 0),
		subCount:    0,
		firstAdd:    false,
		outVADBytes: make([]byte, 0),
	}
}

// AddAndInsert 添加一句对白进来,并且填充中间的空白，间隔 10ms
func (s *SubUnit) AddAndInsert(oneSubStartTime, oneSubEndTime time.Time) {

	/*
		这里有个比较有意思的细节，字幕拆分到 dialogue 的时候，可能连续的多个 dialogue 是时间轴连续的
		但是实际上的语言就是可以分为几个句子的
		那么，在本函数中，就需要判断插入的时候，与上一句话的时间轴关系，前置无需进行句子的合并
		如果两句话时间轴是连续的（差值为0），那么就要主动修改这一点，采取的方案可以是
		1. 前后各 0.001 秒即可
		2. 后面这一句向后 0.002 秒（暂时优先考虑这个，容易实现）
	*/
	const perWindows = float64(vad.FrameDuration) / 1000

	// 不是第一次添加，那么就需要把两句对白中间间隔的 active == false 的插入，插入间隙
	if len(s.VADList) > 0 {
		nowStartTime := s.RealTimeToOffsetTime(oneSubStartTime)
		nowStartOffsetTime := my_util.Time2SecendNumber(nowStartTime)
		nowEndOffsetTime := s.GetEndTimeNumber(false)

		needAddRange := nowStartOffsetTime - nowEndOffsetTime

		if needAddRange == 0 {
			// 说明是连续的句子，向后加 0.002 秒
			addMore := time.Duration((s.GetEndTimeNumber(true) + 0.002) * math.Pow10(9))
			s.VADList = append(s.VADList, *vad.NewVADInfoBase(false, addMore))
			// 因为是连续的两句话的时间轴，强制插入了一个点，那么就需要在这句话的 Start 部分向后延迟对应的秒数
			oneSubStartTime = oneSubStartTime.Add(time.Duration(0.002 * math.Pow10(9)))
		} else {
			for i := 0.0; i < needAddRange; {

				s.VADList = append(s.VADList, *vad.NewVADInfoBase(false, time.Duration((s.GetEndTimeNumber(true)+i)*math.Pow10(9))))
				i += perWindows
			}
		}
	}

	if s.firstAdd == false {
		// 第一次 Add 需要给 baseTime 赋值
		s.baseTime = oneSubStartTime
		s.offsetStartTime = s.RealTimeToOffsetTime(oneSubStartTime)
		s.firstAdd = true
	}

	s.offsetEndTime = oneSubEndTime.Add(-my_util.Time2Duration(s.baseTime))

	nowStartTime := s.RealTimeToOffsetTime(oneSubStartTime)
	nowEndTime := s.RealTimeToOffsetTime(oneSubEndTime)

	nowStartOffsetTime := my_util.Time2SecendNumber(nowStartTime)
	nowEndOffsetTime := my_util.Time2SecendNumber(nowEndTime)

	needAddRange := nowEndOffsetTime - nowStartOffsetTime

	for i := 0.0; i < needAddRange; {
		s.VADList = append(s.VADList, *vad.NewVADInfoBase(true, time.Duration((my_util.Time2SecendNumber(oneSubStartTime)+i)*math.Pow10(9))))
		i += perWindows
	}

	s.subCount++
}

// GetDialogueCount 获取这个对白单元由几个对话
func (s SubUnit) GetDialogueCount() int {
	return s.subCount
}

// GetVADSlice 获取 VAD 的 byte 数组信息
func (s *SubUnit) GetVADSlice() []byte {

	if len(s.outVADBytes) != len(s.VADList) {
		s.outVADBytes = make([]byte, len(s.VADList))
		for i := 0; i < len(s.VADList); i++ {
			if s.VADList[i].Active == true {
				s.outVADBytes[i] = 1
			} else {
				s.outVADBytes[i] = 0
			}
		}
	}

	return s.outVADBytes
}

// GetStartTimeNumber 获取这个单元的起始时间，单位是秒
func (s SubUnit) GetStartTimeNumber(realOrOffsetTime bool) float64 {
	return my_util.Time2SecendNumber(s.GetStartTime(realOrOffsetTime))
}

// GetStartTime 获取这个单元的起始时间
func (s SubUnit) GetStartTime(realOrOffsetTime bool) time.Time {
	if realOrOffsetTime == true {
		return s.offsetStartTime.Add(my_util.Time2Duration(s.baseTime))
	} else {
		return s.offsetStartTime
	}
}

// GetEndTimeNumber 获取这个单元的结束时间，单位是秒
func (s SubUnit) GetEndTimeNumber(realOrOffsetTime bool) float64 {

	return my_util.Time2SecendNumber(s.GetEndTime(realOrOffsetTime))
}

// GetEndTime 获取这个单元的起始时间
func (s SubUnit) GetEndTime(realOrOffsetTime bool) time.Time {
	if realOrOffsetTime == true {
		return s.offsetEndTime.Add(my_util.Time2Duration(s.baseTime))
	} else {
		return s.offsetEndTime
	}
}

// GetTimelineRange 开始到结束的时间长度，单位是秒
func (s SubUnit) GetTimelineRange() float64 {
	return s.GetEndTimeNumber(false) - s.GetStartTimeNumber(false)
}

// GetOffsetTimeNumber 偏移时间，单位是秒
func (s SubUnit) GetOffsetTimeNumber() float64 {
	return my_util.Time2SecendNumber(s.baseTime)
}

// GetFFMPEGCutRange 这里会生成导出 FFMPEG 的参数字段，起始时间和结束的时间长度
func (s SubUnit) GetFFMPEGCutRange(expandTimeRange float64) (string, string) {

	var tmpStartTime time.Time
	if s.GetStartTimeNumber(true)-expandTimeRange < 0 {
		tmpStartTime = time.Time{}
	} else {
		startTime := s.GetStartTime(true)
		subTime := time.Duration(expandTimeRange) * time.Second
		tmpStartTime = startTime.Add(-subTime)
	}

	return fmt.Sprintf("%d:%d:%d.%d", tmpStartTime.Hour(), tmpStartTime.Minute(), tmpStartTime.Second(), tmpStartTime.Nanosecond()/1000/1000),
		fmt.Sprintf("%f", s.GetTimelineRange()+expandTimeRange)
}

// RealTimeToOffsetTime 真实时间转偏移时间
func (s SubUnit) RealTimeToOffsetTime(realTime time.Time) time.Time {
	dd := my_util.Time2Duration(s.baseTime)
	return realTime.Add(-dd)
}
