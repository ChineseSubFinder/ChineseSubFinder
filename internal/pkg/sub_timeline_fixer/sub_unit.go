package sub_timeline_fixer

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/vad"
	"math"
	"time"
)

type SubUnit struct {
	baseTime  time.Time // 这个是基础的时间，后续需要减去这个，不然与导出的片段字幕去对比会有一个起始时间的偏差
	StartTime time.Time // 这个时间会减去 baseTime 再存储
	EndTime   time.Time // 这个时间会减去 baseTime 再存储
	VADList   []vad.VADInfo
	subCount  int
	firstAdd  bool
}

func NewSubUnit() *SubUnit {
	return &SubUnit{
		VADList:  make([]vad.VADInfo, 0),
		subCount: 0,
		firstAdd: false,
	}
}

// Add 添加一句对白进来
func (s *SubUnit) Add(oneSubStartTime, oneSubEndTime time.Time) {

	if s.firstAdd == false {
		s.baseTime = oneSubStartTime
		s.StartTime = oneSubStartTime.Add(-my_util.Time2Duration(s.baseTime))
		s.firstAdd = true
	}
	s.EndTime = oneSubEndTime.Add(-my_util.Time2Duration(s.baseTime))
	// 每一句对白的开始就人为 VAD active 是 1，直到结束，才是 0
	s.VADList = append(s.VADList, *vad.NewVADInfoBase(true, time.Duration(s.GetStartTimeNumber()*math.Pow10(9))))

	s.VADList = append(s.VADList, *vad.NewVADInfoBase(false, time.Duration(s.GetEndTimeNumber()*math.Pow10(9))))

	s.subCount++
}

// AddAndInsert 添加一句对白进来,并且填充中间的空白，间隔 10ms
func (s *SubUnit) AddAndInsert(oneSubStartTime, oneSubEndTime time.Time) {

	perWindows := float64(vad.FrameDuration) / 1000
	// 不是第一次添加，那么就需要把两句对白中间间隔的 active == false 的插入，插入间隙
	if len(s.VADList) > 0 {
		dd := my_util.Time2Duration(s.baseTime)
		tmpSubStartTime := oneSubStartTime.Add(-dd)
		needAddRange := my_util.Time2SecendNumber(tmpSubStartTime) - s.GetEndTimeNumber()
		for i := 0.0; i < needAddRange; {

			s.VADList = append(s.VADList, *vad.NewVADInfoBase(false, time.Duration((s.GetEndTimeNumber()+i)*math.Pow10(9))))
			i += perWindows
		}
	}

	if s.firstAdd == false {
		s.baseTime = oneSubStartTime
		dd := my_util.Time2Duration(s.baseTime)
		s.StartTime = oneSubStartTime.Add(-dd)
		s.firstAdd = true
	}

	s.EndTime = oneSubEndTime.Add(-my_util.Time2Duration(s.baseTime))

	needAddRange := my_util.Time2SecendNumber(oneSubEndTime) - my_util.Time2SecendNumber(oneSubStartTime)
	for i := 0.0; i < needAddRange; {

		s.VADList = append(s.VADList, *vad.NewVADInfoBase(true, time.Duration((s.GetStartTimeNumber()+i)*math.Pow10(9))))
		i += perWindows
	}

	s.subCount++
}

// GetDialogueCount 获取这个对白单元由几个对话
func (s SubUnit) GetDialogueCount() int {
	return s.subCount
}

// GetStartTimeNumber 获取这个单元的起始时间，单位是秒
func (s SubUnit) GetStartTimeNumber() float64 {
	return my_util.Time2SecendNumber(s.StartTime.Add(my_util.Time2Duration(s.baseTime)))
}

// GetEndTimeNumber 获取这个单元的结束时间，单位是秒
func (s SubUnit) GetEndTimeNumber() float64 {
	return my_util.Time2SecendNumber(s.EndTime.Add(my_util.Time2Duration(s.baseTime)))
}

// GetTimelineRange 开始到结束的时间长度，单位是秒
func (s SubUnit) GetTimelineRange() float64 {
	return s.GetEndTimeNumber() - s.GetStartTimeNumber()
}

func (s SubUnit) GetBaseTimeNumber() float64 {
	return my_util.Time2SecendNumber(s.baseTime)
}

// GetFFMPEGCutRange 这里会生成导出 FFMPEG 的参数字段，起始时间和结束的时间长度
func (s SubUnit) GetFFMPEGCutRange(expandTimeRange int) (string, string) {

	var tmpStartTime time.Time
	if s.GetStartTimeNumber()-float64(expandTimeRange)*60 < 0 {
		tmpStartTime = time.Time{}
	} else {
		tmpStartTime = s.StartTime.Add(time.Duration(expandTimeRange) * time.Minute).Add(my_util.Time2Duration(s.baseTime))
	}

	return fmt.Sprintf("%d:%d:%d.%d", tmpStartTime.Hour(), tmpStartTime.Minute(), tmpStartTime.Second(), tmpStartTime.Nanosecond()/1000/1000),
		fmt.Sprintf("%f", s.GetTimelineRange()+float64(expandTimeRange)*60.0)
}
