package sub_timeline_fixer

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/vad"
	"time"
)

type SubUnit struct {
	StartTime time.Time
	EndTime   time.Time
	vadList   []vad.VADInfo
}

func NewSubUnit() *SubUnit {
	return &SubUnit{
		vadList: make([]vad.VADInfo, 0),
	}
}

func (s *SubUnit) Add(oneSubStartTime, oneSubEndTime time.Time) {

	if s.GetStartTimeNumber() == 0 {
		s.StartTime = oneSubStartTime
	}
	s.EndTime = oneSubEndTime
	//
}

func (s SubUnit) GetStartTimeNumber() float64 {
	return pkg.Time2Number(s.StartTime)
}

func (s SubUnit) GetEndTimeNumber() float64 {
	return pkg.Time2Number(s.EndTime)
}

func (s SubUnit) GetFFMPEGCutRange() (string, string) {
	return fmt.Sprintf("%d:%d:%d", s.StartTime.Hour(), s.StartTime.Minute(), s.StartTime.Second()),
		fmt.Sprintf("%f", s.GetEndTimeNumber()-s.GetStartTimeNumber())
}
