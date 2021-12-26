package sub_timeline_fixer

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/gss"
	"github.com/allanpk716/ChineseSubFinder/internal/types/subparser"
)

type Pipeline struct {
	framerateRatios []float64
}

func NewPipeline() *Pipeline {
	return &Pipeline{
		framerateRatios: make([]float64, 0),
	}
}

func (p Pipeline) Fit_gss(infoBase, infoSrc subparser.FileInfo) error {

	/*
		ffsubsync 的 pipeline 有这三个步骤
		1. parse			解析字幕
		2. scale			根据帧数比率调整时间轴
		3. speech_extract	从字幕转换为 VAD 的语音检测信息
	*/
	opt_func := func(framerateRatio float64) float64 {
		nowInfoSrc := infoSrc
		err := nowInfoSrc.ChangeDialoguesFilterExTimeByFramerateRatio(framerateRatio)
		if err != nil {
			// 还原
			println("ChangeDialoguesFilterExTimeByFramerateRatio", err)
			nowInfoSrc = infoSrc
		}
		// 然后进行 base 与 src 匹配计算，将每一次变动 framerateRatio 计算得到的 偏移值和分数进行记录

	}
	gss.Gss(opt_func, MIN_FRAMERATE_RATIO, MAX_FRAMERATE_RATIO, 1e-4, nil)
}

func (p *Pipeline) getFramerateRatios2Try() []float64 {

	if len(p.framerateRatios) > 0 {
		return p.framerateRatios
	}
	p.framerateRatios = append(p.framerateRatios, FRAMERATE_RATIOS...)
	for i := 0; i < len(FRAMERATE_RATIOS); i++ {
		p.framerateRatios = append(p.framerateRatios, 1.0/FRAMERATE_RATIOS[i])
	}
	return p.framerateRatios
}

var FRAMERATE_RATIOS = []float64{24. / 23.976, 25. / 23.976, 25. / 24.}

const MIN_FRAMERATE_RATIO = 0.9
const MAX_FRAMERATE_RATIO = 1.1
