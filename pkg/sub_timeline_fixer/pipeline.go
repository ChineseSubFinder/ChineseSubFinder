package sub_timeline_fixer

import (
	"errors"
	"fmt"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/subparser"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/gss"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/vad"
	"github.com/huandu/go-clone"
)

type Pipeline struct {
	MaxOffsetSeconds int
	framerateRatios  []float64
}

func NewPipeline(maxOffsetSeconds int) *Pipeline {
	return &Pipeline{
		MaxOffsetSeconds: maxOffsetSeconds,
		framerateRatios:  make([]float64, 0),
	}
}

func (p Pipeline) CalcOffsetTime(infoBase, infoSrc *subparser.FileInfo, audioVadList []vad.VADInfo, useGSS bool) (PipeResult, error) {

	baseVADInfo := make([]float64, 0)
	useSubtitleOrAudioAsBase := false
	// 排序
	infoSrc.SortDialogues()
	if infoBase == nil && audioVadList != nil {
		baseVADInfo = vad.GetFloatSlice(audioVadList)
		useSubtitleOrAudioAsBase = true
	} else if infoBase != nil {
		useSubtitleOrAudioAsBase = false
		// 排序
		infoBase.SortDialogues()
		// 解析处 VAD 信息
		baseUnitNew, err := sub_helper.GetVADInfoFeatureFromSubNew(infoBase, 0)
		if err != nil {
			return PipeResult{}, err
		}
		baseVADInfo = baseUnitNew.GetVADFloatSlice()
	} else {
		return PipeResult{}, errors.New("FixTimeline input is error")
	}

	pipeResults := make([]PipeResult, 0)
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
	if useSubtitleOrAudioAsBase == false {
		inferredFramerateRatioFromLength := float64(infoBase.GetNumFrames()) / float64(infoSrc.GetNumFrames())
		framerateRatios = append(framerateRatios, inferredFramerateRatioFromLength)
	}
	// 3.
	fffAligner := NewFFTAligner(p.MaxOffsetSeconds, SampleRate)
	// 需要在这个偏移之下
	maxOffsetSamples := p.MaxOffsetSeconds * SampleRate
	if maxOffsetSamples < 0 {
		maxOffsetSamples = -maxOffsetSamples
	}

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
		err := tmpInfoSrc.ChangeDialoguesTimeByFramerateRatio(framerateRatio)
		if err != nil {
			// 还原
			println("ChangeDialoguesTimeByFramerateRatio", err)
			tmpInfoSrc = clone.Clone(infoSrc).(*subparser.FileInfo)
		}
		// 3. speech_extract	从字幕转换为 VAD 的语音检测信息
		tmpSrcInfoUnit, err := sub_helper.GetVADInfoFeatureFromSubNew(tmpInfoSrc, 0)
		if err != nil {
			return PipeResult{}, err
		}
		bestOffset, score := fffAligner.Fit(baseVADInfo, tmpSrcInfoUnit.GetVADFloatSlice())
		pipeResult := PipeResult{
			Score:          score,
			BestOffset:     bestOffset,
			ScaleFactor:    framerateRatio,
			ScaledFileInfo: tmpInfoSrc,
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
			err := tmpInfoSrc.ChangeDialoguesTimeByFramerateRatio(framerateRatio)
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
			bestOffset, score := fffAligner.Fit(baseVADInfo, tmpSrcInfoUnit.GetVADFloatSlice())
			println(fmt.Sprintf("got score %.0f (offset %d) for ratio %.3f", score, bestOffset, framerateRatio))
			// 放到外部的存储中
			if isLastIter == true {
				pipeResult := PipeResult{
					Score:          score,
					BestOffset:     bestOffset,
					ScaleFactor:    framerateRatio,
					ScaledFileInfo: tmpInfoSrc,
				}
				pipeResults = append(pipeResults, pipeResult)
			}
			return -score
		}

		gss.Gss(optFunc, MinFramerateRatio, MaxFramerateRatio, 1e-4, nil)
	}
	// 先进行过滤
	filterPipeResults := make([]PipeResult, 0)
	for _, result := range pipeResults {
		if result.BestOffset < maxOffsetSamples {
			filterPipeResults = append(filterPipeResults, result)
		}
	}
	if len(filterPipeResults) <= 0 {
		return PipeResult{}, errors.New(fmt.Sprintf("AutoFixTimeline failed; you can set 'MaxOffSetTime' > %d", p.MaxOffsetSeconds) +
			fmt.Sprintf(" Or this two subtiles are not fited to this video!"))
	}
	// 从得到的结果里面找到分数最高的
	sort.Sort(PipeResults(filterPipeResults))
	maxPipeResult := filterPipeResults[len(filterPipeResults)-1]

	return maxPipeResult, nil
}

// FixSubFileTimeline 这里传入的 scaledInfoSrc 是从 pipeResults 筛选出来的最大分数的 FileInfo
// infoSrc 是从源文件读取出来的，这样才能正确匹配 Content 中的时间戳
func (p Pipeline) FixSubFileTimeline(infoSrc, scaledInfoSrc *subparser.FileInfo, inOffsetTime float64, desSaveSubFileFullPath string) (string, error) {

	/*
		从解析的实例中，正常来说是可以匹配出所有的 Dialogue 对话的 Start 和 End time 的信息
		然后找到对应的字幕的文件，进行文件内容的替换来做时间轴的校正
	*/
	// 偏移时间
	offsetTime := time.Duration(inOffsetTime*1000) * time.Millisecond
	fixContent := scaledInfoSrc.Content
	/*
		这里进行时间转字符串的时候有一点比较特殊
		正常来说输出的格式是类似 15:04:05.00
		那么有个问题，字幕的时间格式是 0:00:12.00， 小时，是个数，除非有跨度到 20 小时的视频，不然小时就应该是个数
		这就需要一个额外的函数去处理这些情况
	*/
	timeFormat := scaledInfoSrc.GetTimeFormat()
	// 如果两个解析出来的对白数量不一致，那么肯定是无法进行下面的匹配的，理论上应该没得问题
	if len(scaledInfoSrc.Dialogues) != len(infoSrc.Dialogues) {
		return "", errors.New("FixSubFileTimeline Not The Same Len: scaledInfoSrc.Dialogues and infoSrc.Dialogues")
	}
	contentReplaceOffsetAll := -1
	for index, scaledSrcOneDialogue := range scaledInfoSrc.Dialogues {

		timeStart, err := pkg.ParseTime(scaledSrcOneDialogue.StartTime)
		if err != nil {
			return "", err
		}
		timeEnd, err := pkg.ParseTime(scaledSrcOneDialogue.EndTime)
		if err != nil {
			return "", err
		}

		fixTimeStart := timeStart.Add(offsetTime)
		fixTimeEnd := timeEnd.Add(offsetTime)
		/*
			这里有一个梗（之前没有考虑到），理论上这样的替换应该匹配到一句话（正确的那一句），但是有一定几率
			会把上面修复完的对白时间也算进去替换（匹配上了两句话），导致时间轴无形中被错误延长了
			那么就需要一个 contentReplaceOffsetAll 去记录现在进行到整个字幕那个偏移未知的替换操作了

			并不是说一个字幕中不能出现多个一样的“时间字符串”，也就是如果使用 Find 去查找应该也是一定 >= 1 的结果
			所以才需要 contentReplaceOffsetAll 来记录替换的偏移位置，每次只能变大，而不是变小
		*/

		orgStartTimeString := infoSrc.Dialogues[index].StartTime
		orgEndTimeString := infoSrc.Dialogues[index].EndTime
		// contentReplaceOffsetAll 为 -1 的时候那么第一次搜索得到的就一定是可以替换的
		if contentReplaceOffsetAll == -1 {
			contentReplaceOffsetAll = 0
		}
		contentReplaceOffsetNow := strings.Index(fixContent[contentReplaceOffsetAll:], orgStartTimeString)
		if contentReplaceOffsetNow == -1 {
			// 说明没找到，就跳过,虽然理论上不应该会出现
			continue
		}
		contentReplaceOffsetAll += contentReplaceOffsetNow
		fixContent = fixContent[:contentReplaceOffsetAll] + strings.Replace(fixContent[contentReplaceOffsetAll:], orgStartTimeString, pkg.Time2SubTimeString(fixTimeStart, timeFormat), 1)

		// contentReplaceOffsetAll 为 -1 的时候那么第一次搜索得到的就一定是可以替换的
		if contentReplaceOffsetAll == -1 {
			contentReplaceOffsetAll = 0
		}
		contentReplaceOffsetNow = strings.Index(fixContent[contentReplaceOffsetAll:], orgEndTimeString)
		if contentReplaceOffsetNow == -1 {
			// 说明没找到，就跳过,虽然理论上不应该会出现
			continue
		}
		contentReplaceOffsetAll += contentReplaceOffsetNow
		fixContent = fixContent[:contentReplaceOffsetAll] + strings.Replace(fixContent[contentReplaceOffsetAll:], orgEndTimeString, pkg.Time2SubTimeString(fixTimeEnd, timeFormat), 1)
	}

	dstFile, err := os.Create(desSaveSubFileFullPath)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = dstFile.Close()
	}()
	_, err = dstFile.WriteString(fixContent)
	if err != nil {
		return "", err
	}
	return fixContent, nil
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
const DefaultMaxOffsetSeconds = 120
const SampleRate = 100

type PipeResult struct {
	Score          float64
	BestOffset     int
	ScaleFactor    float64
	ScaledFileInfo *subparser.FileInfo
}

// GetOffsetTime 从偏移得到偏移时间
func (p PipeResult) GetOffsetTime() float64 {
	return float64(p.BestOffset) / 100.0
}

type PipeResults []PipeResult

func (d PipeResults) Len() int {
	return len(d)
}

func (d PipeResults) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

func (d PipeResults) Less(i, j int) bool {

	return d[i].Score < d[j].Score
}
