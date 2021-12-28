package sub_timeline_fixer

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/gss"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/types/subparser"
	"github.com/huandu/go-clone"
)

type Pipeline struct {
	framerateRatios []float64
}

func NewPipeline() *Pipeline {
	return &Pipeline{
		framerateRatios: make([]float64, 0),
	}
}

func (p Pipeline) Fit(infoBase, infoSrc *subparser.FileInfo, useGSS bool) error {

	pipeResults := make([]PipeResult, 0)
	// 排序
	infoBase.SortDialogues()
	infoSrc.SortDialogues()
	println(fmt.Sprintf("%f", my_util.Time2SecondNumber(infoBase.GetStartTime())))
	println(fmt.Sprintf("%f", my_util.Time2SecondNumber(infoBase.GetEndTime())))
	// 解析处 VAD 信息
	baseUnitNew, err := sub_helper.GetVADInfoFeatureFromSubNew(infoBase, 0)
	if err != nil {
		return err
	}
	/*
			这里复现 ffsubsync 的思路
			1. 首先由 getFramerateRatios2Try 得到多个帧数比率的数值，理论上有以下 7 个值：
		       将 frameRateRatio = 1.0 插入到 framerateRatios 这个队列的首位
				[0] 1.0
				[1] 1.001001001001001
				[2] 1.0427093760427095
				[3] 1.0416666666666667
				[4] 0.9989999999999999
				[5] 0.9590399999999999
				[6] 0.96
				得到一个 framerateRatios 列表
			2. 计算 base 字幕的 num_frames，以及 frameRateRatio = 1.0 时 src 字幕的 num_frames
				推断 frame ratio 比率是多少，得到一个，inferred_framerate_ratio_from_length = base / src
				把这个值插入到 framerateRatios 的尾部也就是第八个元素
			3. 使用上述的 framerateRatios 作为传入参数，开始 FFT 模块的 fit 计算，得到（分数、偏移）信息，选择分数最大的作为匹配的结论

	*/
	// 1.
	framerateRatios := make([]float64, 0)
	framerateRatios = p.getFramerateRatios2Try()
	// 2.
	inferredFramerateRatioFromLength := float64(infoBase.GetNumFrames()) / float64(infoSrc.GetNumFrames())
	framerateRatios = append(framerateRatios, inferredFramerateRatioFromLength)
	// 3.
	fffAligner := NewFFTAligner(DefaultMaxOffsetSeconds, SampleRate)
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
		err = tmpInfoSrc.ChangeDialoguesTimeByFramerateRatio(framerateRatio)
		if err != nil {
			// 还原
			println("ChangeDialoguesTimeByFramerateRatio", err)
			tmpInfoSrc = clone.Clone(infoSrc).(*subparser.FileInfo)
		}
		// 3. speech_extract	从字幕转换为 VAD 的语音检测信息
		tmpSrcInfoUnit, err := sub_helper.GetVADInfoFeatureFromSubNew(tmpInfoSrc, 0)
		if err != nil {
			return err
		}
		// 不是用 GSS
		bestOffset, score := fffAligner.Fit(baseUnitNew.GetVADFloatSlice(), tmpSrcInfoUnit.GetVADFloatSlice())
		pipeResult := PipeResult{
			Score:       score,
			BestOffset:  bestOffset,
			ScaleFactor: framerateRatio,
		}
		pipeResults = append(pipeResults, pipeResult)
	}

	if useGSS == true {
		// 最后一个才需要额外使用 GSS
		// 使用 GSS
		optFunc := func(framerateRatio float64, isLastIter bool) float64 {

			// 1. parse			解析字幕
			tmpInfoSrc := clone.Clone(infoSrc).(*subparser.FileInfo)
			// 2. scale			根据帧数比率调整时间轴
			err = tmpInfoSrc.ChangeDialoguesTimeByFramerateRatio(framerateRatio)
			if err != nil {
				// 还原
				println("ChangeDialoguesTimeByFramerateRatio", err)
				tmpInfoSrc = clone.Clone(infoSrc).(*subparser.FileInfo)
			}
			// 3. speech_extract	从字幕转换为 VAD 的语音检测信息
			tmpSrcInfoUnit, err := sub_helper.GetVADInfoFeatureFromSubNew(tmpInfoSrc, 0)
			if err != nil {
				return 0
			}
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
