package sub_timeline_fixer

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/gss"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/types/subparser"
	"github.com/huandu/go-clone"
	"sort"
)

type Pipeline struct {
	framerateRatios []float64
}

func NewPipeline() *Pipeline {
	return &Pipeline{
		framerateRatios: make([]float64, 0),
	}
}

func (p Pipeline) FitGSS(infoBase, infoSrc *subparser.FileInfo) error {

	pipeResults := make([]PipeResult, 0)
	// 排序
	sort.Sort(subparser.OneDialogueByStartTime(infoBase.DialoguesFilter))
	sort.Sort(subparser.OneDialogueByStartTime(infoSrc.DialoguesFilter))
	// 解析处 VAD 信息
	baseUnitNew, err := sub_helper.GetVADInfoFeatureFromSubNew(infoBase, 0)
	if err != nil {
		return err
	}
	fffAligner := NewFFTAligner(DefaultMaxOffsetSeconds, SampleRate)

	framerateRatios := p.getFramerateRatios2Try()
	for _, framerateRatio := range framerateRatios {

		/*
			ffsubsync 的 pipeline 有这三个步骤
			1. parse			解析字幕
			2. scale			根据帧数比率调整时间轴
			3. speech_extract	从字幕转换为 VAD 的语音检测信息
		*/
		// 外部传入
		// 1. parse			解析字幕
		tmpInfoSrc := clone.Clone(infoSrc).(*subparser.FileInfo)
		// 2. scale			根据帧数比率调整时间轴
		err := tmpInfoSrc.ChangeDialoguesFilterExTimeByFramerateRatio(framerateRatio)
		if err != nil {
			// 还原
			println("ChangeDialoguesFilterExTimeByFramerateRatio", err)
			tmpInfoSrc = clone.Clone(infoSrc).(*subparser.FileInfo)
		}
		tmpSrcInfoUnit, err := sub_helper.GetVADInfoFeatureFromSubNew(tmpInfoSrc, 0)
		if err != nil {
			return err
		}

		optFunc := func(framerateRatio float64, isLastIter bool) float64 {

			// 3. speech_extract	从字幕转换为 VAD 的语音检测信息
			// 然后进行 base 与 src 匹配计算，将每一次变动 framerateRatio 计算得到的 偏移值和分数进行记录
			bestOffset, score := fffAligner.Fit(baseUnitNew.GetVADFloatSlice(), tmpSrcInfoUnit.GetVADFloatSlice())
			println(fmt.Sprintf("got score %.0f (offset %d) for ratio %.3f", score, bestOffset, framerateRatio))
			// 放到外部的存储中
			if isLastIter == true {
				pipeResult := PipeResult{
					Score:       score,
					BestOffset:  bestOffset,
					ScaleFactor: framerateRatio,
				}
				pipeResults = append(pipeResults, pipeResult)
			}
			return -score
		}

		gss.Gss(optFunc, MinFramerateRatio, MaxFramerateRatio, 1e-4, nil)
	}

	return nil
}

func (p *Pipeline) getFramerateRatios2Try() []float64 {

	if len(p.framerateRatios) > 0 {
		return p.framerateRatios
	}
	p.framerateRatios = append(p.framerateRatios, 1.0)
	p.framerateRatios = append(p.framerateRatios, FramerateRatios...)
	for i := 0; i < len(FramerateRatios); i++ {
		p.framerateRatios = append(p.framerateRatios, 1.0/FramerateRatios[i])
	}
	return p.framerateRatios
}

var FramerateRatios = []float64{24. / 23.976, 25. / 23.976, 25. / 24.}

const MinFramerateRatio = 0.9
const MaxFramerateRatio = 1.1
const DefaultMaxOffsetSeconds = 60
const SampleRate = 100

type PipeResult struct {
	Score       float64
	BestOffset  int
	ScaleFactor float64
}
