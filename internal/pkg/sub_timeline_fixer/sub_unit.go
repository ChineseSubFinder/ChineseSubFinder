package sub_timeline_fixer

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/vad"
	"math"
	"time"
)

type SubUnit struct {
	StartTime time.Time
	EndTime   time.Time
	vadList   []vad.VADInfo
	subCount  int
}

func NewSubUnit() *SubUnit {
	return &SubUnit{
		vadList:  make([]vad.VADInfo, 0),
		subCount: 0,
	}
}

// Add 添加一句对白进来
func (s *SubUnit) Add(oneSubStartTime, oneSubEndTime time.Time) {

	if s.GetStartTimeNumber() == 0 {
		s.StartTime = oneSubStartTime
	}
	s.EndTime = oneSubEndTime
	// 每一句对白的开始就人为 VAD active 是 1，直到结束，才是 0
	s.vadList = append(s.vadList, *vad.NewVADInfoBase(true, time.Duration(s.GetStartTimeNumber()*math.Pow10(9))))

	s.vadList = append(s.vadList, *vad.NewVADInfoBase(false, time.Duration(s.GetEndTimeNumber()*math.Pow10(9))))

	s.subCount++
}

// AddAndInsert 添加一句对白进来,并且填充中间的空白，间隔 10ms
func (s *SubUnit) AddAndInsert(oneSubStartTime, oneSubEndTime time.Time) {

	perWindows := float64(vad.FrameDuration) / 1000
	// 不是第一次添加，那么就需要把两句对白中间间隔的 active == false 的插入，插入间隙
	if len(s.vadList) > 0 {
		needAddRange := my_util.Time2SecendNumber(oneSubStartTime) - s.GetEndTimeNumber()
		for i := 0.0; i < needAddRange; {

			s.vadList = append(s.vadList, *vad.NewVADInfoBase(false, time.Duration((s.GetEndTimeNumber()+i)*math.Pow10(9))))
			i += perWindows
		}
	}

	if s.GetStartTimeNumber() == 0 {
		s.StartTime = oneSubStartTime
	}
	s.EndTime = oneSubEndTime

	needAddRange := my_util.Time2SecendNumber(oneSubEndTime) - my_util.Time2SecendNumber(oneSubStartTime)
	for i := 0.0; i < needAddRange; {

		s.vadList = append(s.vadList, *vad.NewVADInfoBase(true, time.Duration((s.GetStartTimeNumber()+i)*math.Pow10(9))))
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
	return my_util.Time2SecendNumber(s.StartTime)
}

// GetEndTimeNumber 获取这个单元的结束时间，单位是秒
func (s SubUnit) GetEndTimeNumber() float64 {
	return my_util.Time2SecendNumber(s.EndTime)
}

// GetTimelineRange 开始到结束的时间长度，单位是秒
func (s SubUnit) GetTimelineRange() float64 {
	return s.GetEndTimeNumber() - s.GetStartTimeNumber()
}

// GetFFMPEGCutRange 这里会生成导出 FFMPEG 的参数字段，起始时间和结束的时间长度
func (s SubUnit) GetFFMPEGCutRange(expandTimeRange int) (string, string) {

	var tmpStartTime time.Time
	if s.GetStartTimeNumber()-float64(expandTimeRange)*60 < 0 {
		tmpStartTime = time.Time{}
	} else {
		tmpStartTime = s.StartTime.Add(time.Duration(expandTimeRange) * time.Minute)
	}

	return fmt.Sprintf("%d:%d:%d.%d", tmpStartTime.Hour(), tmpStartTime.Minute(), tmpStartTime.Second(), tmpStartTime.Nanosecond()/1000/1000),
		fmt.Sprintf("%f", s.GetTimelineRange()+float64(expandTimeRange)*60.0)
}
