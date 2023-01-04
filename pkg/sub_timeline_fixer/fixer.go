package sub_timeline_fixer

//
//import (
//	"errors"
//	"fmt"
//	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/log_helper"
//	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/my_util"
//	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_helper"
//	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/vad"
//	"github.com/ChineseSubFinder/ChineseSubFinder/internal/types/sub_timeline_fiexer"
//	"github.com/ChineseSubFinder/ChineseSubFinder/internal/types/subparser"
//	"github.com/emirpasic/gods/maps/treemap"
//	"github.com/emirpasic/gods/utils"
//	"github.com/go-echarts/go-echarts/v2/opts"
//	"github.com/grd/stat"
//	"github.com/james-bowman/nlp/measures/pairwise"
//	"github.com/mndrix/tukey"
//	"github.com/panjf2000/ants/v2"
//	"golang.org/x/net/context"
//	"gonum.org/v1/gonum/mat"
//	"math"
//	"os"
//	"sort"
//	"strings"
//	"sync"
//	"time"
//)
//
//type SubTimelineFixer struct {
//  log *logrus.Logger
//	FixerConfig sub_timeline_fiexer.SubTimelineFixerConfig
//}
//
//func NewSubTimelineFixer(log *logrus.Logger, fixerConfig sub_timeline_fiexer.SubTimelineFixerConfig) *SubTimelineFixer {
//	return &SubTimelineFixer{
//		log: log,
//		FixerConfig: fixerConfig,
//	}
//}
//
//// StopWordCounter 停止词统计
//func (s *SubTimelineFixer) StopWordCounter(inString string, per int) []string {
//	statisticTimes := make(map[string]int)
//	wordsLength := strings.Fields(inString)
//
//	for counts, word := range wordsLength {
//		// 判断key是否存在，这个word是字符串，这个counts是统计的word的次数。
//		word, ok := statisticTimes[word]
//		if ok {
//			word = word
//			statisticTimes[wordsLength[counts]] = statisticTimes[wordsLength[counts]] + 1
//		} else {
//			statisticTimes[wordsLength[counts]] = 1
//		}
//	}
//
//	stopWords := make([]string, 0)
//	mapByValue := sortMapByValue(statisticTimes)
//
//	breakIndex := len(mapByValue) * per / 100
//	for index, wordInfo := range mapByValue {
//		if index > breakIndex {
//			break
//		}
//		stopWords = append(stopWords, wordInfo.Name)
//	}
//
//	return stopWords
//}
//
//// FixSubTimelineOneOffsetTime 校正整个字幕文件的时间轴，适用于一个偏移值的情况
//func (s *SubTimelineFixer) FixSubTimelineOneOffsetTime(infoSrc *subparser.FileInfo, inOffsetTime float64, desSaveSubFileFullPath string) (string, error) {
//
//	/*
//		从解析的实例中，正常来说是可以匹配出所有的 Dialogue 对话的 Start 和 End time 的信息
//		然后找到对应的字幕的文件，进行文件内容的替换来做时间轴的校正
//	*/
//	// 偏移时间
//	offsetTime := time.Duration(inOffsetTime*1000) * time.Millisecond
//	fixContent := infoSrc.Content
//	/*
//		这里进行时间转字符串的时候有一点比较特殊
//		正常来说输出的格式是类似 15:04:05.00
//		那么有个问题，字幕的时间格式是 0:00:12.00， 小时，是个数，除非有跨度到 20 小时的视频，不然小时就应该是个数
//		这就需要一个额外的函数去处理这些情况
//	*/
//	timeFormat := infoSrc.GetTimeFormat()
//	for _, srcOneDialogue := range infoSrc.Dialogues {
//
//		timeStart, err := my_util.ParseTime(srcOneDialogue.StartTime)
//		if err != nil {
//			return "", err
//		}
//		timeEnd, err := my_util.ParseTime(srcOneDialogue.EndTime)
//		if err != nil {
//			return "", err
//		}
//
//		fixTimeStart := timeStart.Add(offsetTime)
//		fixTimeEnd := timeEnd.Add(offsetTime)
//
//		fixContent = strings.ReplaceAll(fixContent, srcOneDialogue.StartTime, my_util.Time2SubTimeString(fixTimeStart, timeFormat))
//		fixContent = strings.ReplaceAll(fixContent, srcOneDialogue.EndTime, my_util.Time2SubTimeString(fixTimeEnd, timeFormat))
//	}
//
//	dstFile, err := os.Create(desSaveSubFileFullPath)
//	if err != nil {
//		return "", err
//	}
//	defer func() {
//		_ = dstFile.Close()
//	}()
//	_, err = dstFile.WriteString(fixContent)
//	if err != nil {
//		return "", err
//	}
//	return fixContent, nil
//}
//
//// FixSubTimelineByFixResults V2 专用的时间校正函数
//func (s SubTimelineFixer) FixSubTimelineByFixResults(infoSrc *subparser.FileInfo, srcUnitNew *sub_helper.SubUnit, fixedResults []FixResult, desSaveSubFileFullPath string) (string, error) {
//
//	startTime := srcUnitNew.GetStartTime(true)
//	startTimeBaseDouble := my_util.Time2SecondNumber(startTime)
//	/*
//		这里拿到的 fixedResults ，是进行过 V2_FrontAndEndPerSrc 头尾去除
//		那么调整目标字幕的时候，需要考虑截取掉的部分也要算进去
//	*/
//	/*
//		从解析的实例中，正常来说是可以匹配出所有的 Dialogue 对话的 Start 和 End time 的信息
//		然后找到对应的字幕的文件，进行文件内容的替换来做时间轴的校正
//	*/
//	fixContent := infoSrc.Content
//	/*
//		这里进行时间转字符串的时候有一点比较特殊
//		正常来说输出的格式是类似 15:04:05.00
//		那么有个问题，字幕的时间格式是 0:00:12.00， 小时，是个数，除非有跨度到 20 小时的视频，不然小时就应该是个数
//		这就需要一个额外的函数去处理这些情况
//	*/
//	timeFormat := infoSrc.GetTimeFormat()
//	cacheIndex := 0
//	/*
//		这里的理想情况是 Dialogues，每一句话都是递增的对白时间
//		但是实际情况可能是，前面几个对白是特效、音乐的备注，那么他们的跨度可以很大
//		然后才到正常的对话对白，这样就出现不是递增的时间对白情况
//		那么就需要对 Dialogues 进行排序，然后再进行处理
//	*/
//	sort.Sort(subparser.OneDialogueByStartTime(infoSrc.Dialogues))
//
//	for index, srcOneDialogue := range infoSrc.Dialogues {
//
//		timeStart, err := my_util.ParseTime(srcOneDialogue.StartTime)
//		if err != nil {
//			return "", err
//		}
//		timeEnd, err := my_util.ParseTime(srcOneDialogue.EndTime)
//		if err != nil {
//			return "", err
//		}
//
//		inOffsetTime := 0.0
//		orgStartTimeDouble := my_util.Time2SecondNumber(timeStart)
//		for cacheIndex < len(fixedResults) {
//
//			inRange, nowOffsetTime := fixedResults[cacheIndex].InRange(startTimeBaseDouble, orgStartTimeDouble)
//			if inRange == false {
//				// 大于当前的范围，递增一个区间进行再次的判断
//				// 但是需要确定的是，递增出来的这个区间的 Index 是有效的，如果是无效的，那么就使用最后一个区间的偏移时间
//				cacheIndex++
//				continue
//			} else {
//				inOffsetTime = nowOffsetTime
//				break
//			}
//		}
//		if cacheIndex >= len(fixedResults) {
//			// 下一个区间的 Index 已经越界了，那么就使用最后一个区间的偏移
//			inOffsetTime = fixedResults[len(fixedResults)-1].NewMean
//		}
//		// 偏移时间
//		println(index, inOffsetTime)
//		offsetTime := time.Duration(inOffsetTime*1000) * time.Millisecond
//		fixTimeStart := timeStart.Add(offsetTime)
//		fixTimeEnd := timeEnd.Add(offsetTime)
//
//		fixContent = strings.ReplaceAll(fixContent, srcOneDialogue.StartTime, "Index:"+fmt.Sprintf("%d-", index)+my_util.Time2SubTimeString(fixTimeStart, timeFormat))
//		fixContent = strings.ReplaceAll(fixContent, srcOneDialogue.EndTime, my_util.Time2SubTimeString(fixTimeEnd, timeFormat))
//	}
//
//	dstFile, err := os.Create(desSaveSubFileFullPath)
//	if err != nil {
//		return "", err
//	}
//	defer func() {
//		_ = dstFile.Close()
//	}()
//	_, err = dstFile.WriteString(fixContent)
//	if err != nil {
//		return "", err
//	}
//	return fixContent, nil
//}
//
///*
//	对于 V1 版本的字幕时间轴校正来说，是有特殊的前置要求的
//	1. 视频要有英文字幕
//	2. 外置的字幕必须是中文的双语字幕（简英、繁英）
//*/
//// GetOffsetTimeV1 暂时只支持英文的基准字幕，源字幕必须是双语中英字幕
//func (s *SubTimelineFixer) GetOffsetTimeV1(infoBase, infoSrc *subparser.FileInfo, staticLineFileSavePath string, debugInfoFileSavePath string) (bool, float64, float64, error) {
//
//	var debugInfos = make([]string, 0)
//	// 构建基准语料库，目前阶段只需要考虑是 En 的就行了
//	var baseCorpus = make([]string, 0)
//	var baseDialogueFilterMap = make(map[int]int, 0)
//	/*
//		这里原来的写法是所有的 base 的都放进去匹配，这样会带来一些不必要的对白
//		需要剔除空白。那么就需要建立一个转换的字典
//	*/
//	for index, oneDialogueEx := range infoBase.DialoguesFilterEx {
//		if oneDialogueEx.EnLine == "" {
//			continue
//		}
//		baseCorpus = append(baseCorpus, oneDialogueEx.EnLine)
//		baseDialogueFilterMap[len(baseCorpus)-1] = index
//	}
//	// 初始化
//	pipLine, tfidf, err := NewTFIDF(baseCorpus)
//	if err != nil {
//		return false, 0, 0, err
//	}
//
//	/*
//		确认两个字幕间的偏移，暂定的方案是两边都连续匹配上 5 个索引，再抽取一个对话的时间进行修正计算
//	*/
//	maxCompareDialogue := s.FixerConfig.V1_MaxCompareDialogue
//	// 基线的长度
//	_, docsLength := tfidf.Dims()
//	var matchIndexList = make([]MatchIndex, 0)
//	sc := NewSubCompare(maxCompareDialogue)
//	// 开始比较相似度，默认认为是 Ch_en 就行了
//	for srcIndex := 0; srcIndex < len(infoSrc.DialoguesFilterEx); {
//
//		srcOneDialogueEx := infoSrc.DialoguesFilterEx[srcIndex]
//		// 这里只考虑 英文 的语言
//		if srcOneDialogueEx.EnLine == "" {
//			srcIndex++
//			continue
//		}
//		// run the query through the same pipeline that was fitted to the corpus and
//		// to project it into the same dimensional space
//		queryVector, err := pipLine.Transform(srcOneDialogueEx.EnLine)
//		if err != nil {
//			return false, 0, 0, err
//		}
//		// iterate over document feature vectors (columns) in the LSI matrix and compare
//		// with the query vector for similarity.  Similarity is determined by the difference
//		// between the angles of the vectors known as the cosine similarity
//		highestSimilarity := -1.0
//		// 匹配上的基准的索引
//		var baseIndex int
//		// 这里理论上需要把所有的基线遍历一次，但是，一般来说，两个字幕不可能差距在 50 行
//		// 这样的好处是有助于提高搜索的性能
//		// 那么就以当前的 src 的位置，向前、向后各 50 来遍历
//		nowMaxScanLength := srcIndex + 50
//		nowMinScanLength := srcIndex - 50
//		if nowMinScanLength < 0 {
//			nowMinScanLength = 0
//		}
//		if nowMaxScanLength > docsLength {
//			nowMaxScanLength = docsLength
//		}
//		for i := nowMinScanLength; i < nowMaxScanLength; i++ {
//			similarity := pairwise.CosineSimilarity(queryVector.(mat.ColViewer).ColView(0), tfidf.(mat.ColViewer).ColView(i))
//			if similarity > highestSimilarity {
//				baseIndex = i
//				highestSimilarity = similarity
//			}
//		}
//
//		startBaseIndex, startSrcIndex := sc.GetStartIndex()
//		if sc.Add(baseIndex, srcIndex) == false {
//			sc.Clear()
//			srcIndex = startSrcIndex + 1
//			continue
//			//sc.Add(baseIndex, srcIndex)
//		}
//		if sc.Check() == false {
//			srcIndex++
//			continue
//		} else {
//			sc.Clear()
//		}
//
//		matchIndexList = append(matchIndexList, MatchIndex{
//			BaseNowIndex: startBaseIndex,
//			//BaseNowIndex: baseDialogueFilterMap[startBaseIndex],
//			SrcNowIndex: startSrcIndex,
//			Similarity:  highestSimilarity,
//		})
//
//		//println(fmt.Sprintf("Similarity: %f Base[%d] %s-%s '%s' <--> Src[%d] %s-%s '%s'",
//		//	highestSimilarity,
//		//	baseIndex, infoBase.DialoguesFilterEx[baseIndex].relativelyStartTime, infoBase.DialoguesFilterEx[baseIndex].relativelyEndTime, baseCorpus[baseIndex],
//		//	srcIndex, srcOneDialogueEx.relativelyStartTime, srcOneDialogueEx.relativelyEndTime, srcOneDialogueEx.EnLine))
//
//		srcIndex++
//	}
//
//	var startDiffTimeLineData = make([]opts.LineData, 0)
//	var endDiffTimeLineData = make([]opts.LineData, 0)
//	var tmpStartDiffTime = make([]float64, 0)
//	var tmpEndDiffTime = make([]float64, 0)
//	var startDiffTimeList = make(stat.Float64Slice, 0)
//	var endDiffTimeList = make(stat.Float64Slice, 0)
//	var xAxis = make([]string, 0)
//	// 上面找出了连续匹配 maxCompareDialogue：N 次的字幕语句块
//	// 求出平均时间偏移
//	for mIndex, matchIndexItem := range matchIndexList {
//
//		for i := 0; i < maxCompareDialogue; i++ {
//			// 这里会统计连续的这 5 句话的时间差
//			//tmpBaseIndex := matchIndexItem.BaseNowIndex + i
//			tmpBaseIndex := baseDialogueFilterMap[matchIndexItem.BaseNowIndex+i]
//			tmpSrcIndex := matchIndexItem.SrcNowIndex + i
//
//			baseTimeStart, err := my_util.ParseTime(infoBase.DialoguesFilterEx[tmpBaseIndex].StartTime)
//			if err != nil {
//				return false, 0, 0, err
//			}
//			baseTimeEnd, err := my_util.ParseTime(infoBase.DialoguesFilterEx[tmpBaseIndex].EndTime)
//			if err != nil {
//				return false, 0, 0, err
//			}
//			srtTimeStart, err := my_util.ParseTime(infoSrc.DialoguesFilterEx[tmpSrcIndex].StartTime)
//			if err != nil {
//				return false, 0, 0, err
//			}
//			srtTimeEnd, err := my_util.ParseTime(infoSrc.DialoguesFilterEx[tmpSrcIndex].EndTime)
//			if err != nil {
//				return false, 0, 0, err
//			}
//
//			TimeDiffStart := baseTimeStart.Sub(srtTimeStart)
//			TimeDiffEnd := baseTimeEnd.Sub(srtTimeEnd)
//
//			startDiffTimeLineData = append(startDiffTimeLineData, opts.LineData{Value: TimeDiffStart.Seconds()})
//			endDiffTimeLineData = append(endDiffTimeLineData, opts.LineData{Value: TimeDiffEnd.Seconds()})
//
//			tmpStartDiffTime = append(tmpStartDiffTime, TimeDiffStart.Seconds())
//			tmpEndDiffTime = append(tmpEndDiffTime, TimeDiffEnd.Seconds())
//
//			startDiffTimeList = append(startDiffTimeList, TimeDiffStart.Seconds())
//			endDiffTimeList = append(endDiffTimeList, TimeDiffEnd.Seconds())
//
//			xAxis = append(xAxis, fmt.Sprintf("%d_%d", mIndex, i))
//
//			debugInfos = append(debugInfos, "bs "+infoBase.DialoguesFilterEx[tmpBaseIndex].StartTime+" <-> "+infoBase.DialoguesFilterEx[tmpBaseIndex].EndTime)
//			debugInfos = append(debugInfos, "sc "+infoSrc.DialoguesFilterEx[tmpSrcIndex].StartTime+" <-> "+infoSrc.DialoguesFilterEx[tmpSrcIndex].EndTime)
//			debugInfos = append(debugInfos, "StartDiffTime: "+fmt.Sprintf("%f", TimeDiffStart.Seconds()))
//			//println(fmt.Sprintf("Diff Start-End: %s - %s Base[%d] %s-%s '%s' <--> Src[%d] %s-%s '%s'",
//			//	TimeDiffStart, TimeDiffEnd,
//			//	tmpBaseIndex, infoBase.DialoguesFilterEx[tmpBaseIndex].relativelyStartTime, infoBase.DialoguesFilterEx[tmpBaseIndex].relativelyEndTime, infoBase.DialoguesFilterEx[tmpBaseIndex].EnLine,
//			//	tmpSrcIndex, infoSrc.DialoguesFilterEx[tmpSrcIndex].relativelyStartTime, infoSrc.DialoguesFilterEx[tmpSrcIndex].relativelyEndTime, infoSrc.DialoguesFilterEx[tmpSrcIndex].EnLine))
//		}
//		debugInfos = append(debugInfos, "---------------------------------------------")
//		//println("---------------------------------------------")
//	}
//
//	oldMean := stat.Mean(startDiffTimeList)
//	oldSd := stat.Sd(startDiffTimeList)
//	newMean := -1.0
//	newSd := -1.0
//	per := 1.0
//
//	// 如果 SD 较大的时候才需要剔除
//	if oldSd > 0.1 {
//		var outliersMap = make(map[float64]int, 0)
//		outliers, _, _ := tukey.Outliers(0.3, tmpStartDiffTime)
//		for _, outlier := range outliers {
//			outliersMap[outlier] = 0
//		}
//		var newStartDiffTimeList = make([]float64, 0)
//		for _, f := range tmpStartDiffTime {
//
//			_, ok := outliersMap[f]
//			if ok == true {
//				continue
//			}
//
//			newStartDiffTimeList = append(newStartDiffTimeList, f)
//		}
//
//		orgLen := startDiffTimeList.Len()
//		startDiffTimeList = make(stat.Float64Slice, 0)
//		for _, f := range newStartDiffTimeList {
//			startDiffTimeList = append(startDiffTimeList, f)
//		}
//		newLen := startDiffTimeList.Len()
//
//		per = float64(newLen) / float64(orgLen)
//
//		newMean = stat.Mean(startDiffTimeList)
//		newSd = stat.Sd(startDiffTimeList)
//	}
//
//	if newMean == -1.0 {
//		newMean = oldMean
//	}
//	if newSd == -1.0 {
//		newSd = oldSd
//	}
//
//	// 不为空的时候，生成调试文件
//	if staticLineFileSavePath != "" {
//		//staticLineFileSavePath = "bar.html"
//		err = SaveStaticLineV1(staticLineFileSavePath, infoBase.Name, infoSrc.Name,
//			per, oldMean, oldSd, newMean, newSd, xAxis,
//			startDiffTimeLineData, endDiffTimeLineData)
//		if err != nil {
//			return false, 0, 0, err
//		}
//	}
//
//	// 跳过的逻辑是 mean 是 0 ，那么现在如果判断有问题，缓存的调试文件继续生成，然后强制返回 0 来跳过后续的逻辑
//	// 这里需要考虑，找到的连续 5 句话匹配的有多少句，占比整体所有的 Dialogue 是多少，太低也需要跳过
//	matchIndexLineCount := len(matchIndexList) * maxCompareDialogue
//	//perMatch := float64(matchIndexLineCount) / float64(len(infoSrc.DialoguesFilterEx))
//	perMatch := float64(matchIndexLineCount) / float64(len(baseCorpus))
//	if perMatch < s.FixerConfig.V1_MinMatchedPercent {
//		tmpContent := infoSrc.Name + fmt.Sprintf(" Sequence match %d dialogues (< %f%%), Skip,", s.FixerConfig.V1_MaxCompareDialogue, s.FixerConfig.V1_MinMatchedPercent*100) + fmt.Sprintf(" %f%% ", perMatch*100)
//
//		debugInfos = append(debugInfos, tmpContent)
//
//		s.log.Infoln(tmpContent)
//	} else {
//		tmpContent := infoSrc.Name + fmt.Sprintf(" Sequence match %d dialogues,", s.FixerConfig.V1_MaxCompareDialogue) + fmt.Sprintf(" %f%% ", perMatch*100)
//
//		debugInfos = append(debugInfos, tmpContent)
//
//		s.log.Infoln(tmpContent)
//	}
//
//	// 输出调试的匹配时间轴信息的列表
//	if debugInfoFileSavePath != "" {
//		err = my_util.WriteStrings2File(debugInfoFileSavePath, debugInfos)
//		if err != nil {
//			return false, 0, 0, err
//		}
//	}
//	// 虽然有条件判断是认为有问题的，但是返回值还是要填写除去的
//	if perMatch < s.FixerConfig.V1_MinMatchedPercent {
//		return false, newMean, newSd, nil
//	}
//
//	return true, newMean, newSd, nil
//}
//
//// GetOffsetTimeV2 使用内置的字幕校正外置的字幕时间轴
//func (s *SubTimelineFixer) GetOffsetTimeV2(baseUnit, srcUnit *sub_helper.SubUnit, audioVadList []vad.VADInfo) (bool, []FixResult, error) {
//
//	// -------------------------------------------------
//	/*
//		开始针对对白单元进行匹配
//		下面的逻辑需要参考 FFT识别流程.jpg 这个图示来理解
//		实际实现的时候，会在上述 srcUnit 上，做一个滑动窗口来做匹配，80% 是窗口，20% 用于移动
//		步长固定在 10 步
//	*/
//	audioFloatList := vad.GetFloatSlice(audioVadList)
//
//	srcVADLen := len(srcUnit.VADList)
//	// 滑动窗口的长度
//	srcWindowLen := int(float64(srcVADLen) * s.FixerConfig.V2_WindowMatchPer)
//	// 划分为 4 个区域，每一个部分的长度
//	const parts = 10
//	perPartLen := srcVADLen / parts
//	matchedInfos := make([]MatchInfo, 0)
//
//	subVADBlockInfos := make([]SubVADBlockInfo, 0)
//	for i := 0; i < parts; i++ {
//
//		// 滑动窗体的起始 Index
//		srcSlideStartIndex := i * perPartLen
//		// 滑动的距离
//		srcSlideLen := perPartLen
//		// 一步的长度
//		oneStep := perPartLen / s.FixerConfig.V2_CompareParts
//		if srcSlideLen <= 0 {
//			srcSlideLen = 1
//		}
//		if oneStep <= 0 {
//			oneStep = 1
//		}
//		// -------------------------------------------------
//		windowInfo := WindowInfo{
//			BaseAudioFloatList: audioFloatList,
//			BaseUnit:           baseUnit,
//			SrcUnit:            srcUnit,
//			MatchedTimes:       0,
//			SrcWindowLen:       srcWindowLen,
//			SrcSlideStartIndex: srcSlideStartIndex,
//			SrcSlideLen:        srcSlideLen,
//			OneStep:            oneStep,
//		}
//		subVADBlockInfos = append(subVADBlockInfos, SubVADBlockInfo{
//			Index:      i,
//			StartIndex: srcSlideStartIndex,
//			EndIndex:   srcSlideStartIndex + srcSlideLen,
//		})
//		// 实际 FFT 的匹配逻辑函数
//		// 时间轴差值数组
//		matchInfo, err := s.slidingWindowProcessorV2(&windowInfo)
//		if err != nil {
//			return false, nil, err
//		}
//
//		matchedInfos = append(matchedInfos, *matchInfo)
//	}
//
//	fixedResults := make([]FixResult, 0)
//	sdLessCount := 0
//	// 这里的是 matchedInfos 是顺序的
//	for index, matchInfo := range matchedInfos {
//
//		s.log.Infoln(index, "------------------------------------")
//		outCorrelationFixResult := s.calcMeanAndSDV2(matchInfo.StartDiffTimeListEx, matchInfo.StartDiffTimeList)
//		s.log.Infoln(fmt.Sprintf("FFTAligner Old Mean: %v SD: %f Per: %v", outCorrelationFixResult.OldMean, outCorrelationFixResult.OldSD, outCorrelationFixResult.Per))
//		s.log.Infoln(fmt.Sprintf("FFTAligner New Mean: %v SD: %f Per: %v", outCorrelationFixResult.NewMean, outCorrelationFixResult.NewSD, outCorrelationFixResult.Per))
//
//		value, indexMax := matchInfo.StartDiffTimeMap.Max()
//		s.log.Infoln("FFTAligner Max score:", fmt.Sprintf("%v", value.(float64)), "Time:", fmt.Sprintf("%v", matchInfo.StartDiffTimeList[indexMax.(int)]))
//
//		outCorrelationFixResult.StartVADIndex = index * perPartLen
//		outCorrelationFixResult.EndVADIndex = index*perPartLen + perPartLen
//		fixedResults = append(fixedResults, outCorrelationFixResult)
//
//		if outCorrelationFixResult.NewSD < 0.1 {
//			sdLessCount++
//		}
//	}
//
//	// 如果 0.1 sd 以下的占比低于 70% 那么就认为字幕匹配失败
//	perLess := float64(sdLessCount) / float64(len(matchedInfos))
//	if perLess < 0.7 {
//		return false, nil, nil
//	}
//
//	// matchedInfos 与 fixedResults 是对等的关系，fixedResults 中是计算过 Mean 的值，而 matchedInfos 有原始的值
//	for i, info := range matchedInfos {
//		for j := 0; j < len(info.IndexMatchWindowInfoMap); j++ {
//
//			value, bFound := info.IndexMatchWindowInfoMap[j]
//			if bFound == false {
//				continue
//			}
//
//			fixedResults[i].MatchWindowInfos = append(fixedResults[i].MatchWindowInfos, value)
//		}
//	}
//	/*
//		如果 outCorrelationFixResult 的 SD > 0.1，那么大概率这个时间轴的值匹配的有问题，需要向左或者向右找一个值进行继承
//		-4 0.001
//		-4 0.001
//		-4 0.001
//		-200 0.1
//		-4 0.001
//		比如这种情况，那么就需要向左找到 -4 去继承。
//		具体的实现：
//			找到一个 SD > 0.1 的项目，那么就需要从左边和右边同时对比
//			首先是他们的差值要在 0.3s （绝对值）以内，优先往左边找，如果绝对值成立则判断 SD （SD 必须 < 0.1）
//			如果只是 SD 不成立，那么就继续往左，继续判断差值和 SD。
//			如果都找不到合适的，就要回到”起点“，从右开始找，逻辑一样
//			直到没有找到合适的信息，就报错
//	*/
//	// 进行细节的修正
//	for index, fixedResult := range fixedResults {
//		// SD 大于 0.1 或者是 当前的 NewMean 与上一个点的 NewMean 差值大于 0.3
//		if fixedResult.NewSD >= 0.1 || (index > 1 && math.Abs(fixedResult.NewMean-fixedResults[index-1].NewMean) > 0.3) {
//			bok, newMean, newSD := s.fixOnePartV2(index, fixedResults)
//			if bok == true {
//				fixedResults[index].NewMean = newMean
//				fixedResults[index].NewSD = newSD
//			}
//		}
//	}
//
//	return true, fixedResults, nil
//}
//
//// fixOnePartV2 轻微地跳动可以根据左或者右去微调
//func (s SubTimelineFixer) fixOnePartV2(startIndex int, fixedResults []FixResult) (bool, float64, float64) {
//
//	/*
//		找到这样情况的进行修正
//	*/
//	// 先往左
//	if startIndex-1 >= 0 {
//		// 说明至少可以往左
//		// 如果左边的这个值，与当前值超过了 0.3 的绝对差值，那么是不适合的，就需要往右找
//		if math.Abs(fixedResults[startIndex-1].NewMean-fixedResults[startIndex].NewMean) < 0.3 {
//			// 差值在接受的范围内，那么就使用这个左边的值去校正当前的值
//			return true, fixedResults[startIndex-1].NewMean, fixedResults[startIndex-1].NewSD
//		}
//	}
//
//	// 如果上面的理想情况都没有进去，那么就是这个差值很大
//	if fixedResults[startIndex].NewSD > 1 {
//		// SD 比较大，可能当前的位置是值是错误的，那么直接就使用左边的值
//		/*
//			-6.3	0.06
//			-146.85	243.83
//		*/
//		if startIndex-1 >= 0 {
//			return true, fixedResults[startIndex-1].NewMean, fixedResults[startIndex-1].NewSD
//		}
//	} else {
//		// SD 不是很大，可能就是正常的字幕分段的时间轴偏移的 越接处 !
//		// 那么需要取，越接处，前三和后三，进行均值计算
//		/*
//			-6.21
//			-6.22
//			-6.29	0.06
//
//			-7.13	0.14		越接处
//
//			-7.32
//			-7.31
//			-7.44
//		*/
//		left3Mean := 0.0
//		right3Mean := 0.0
//		// 向左，三个或者三个位置
//		if startIndex-3 >= 0 {
//			left3Mean = float64(fixedResults[startIndex-1].NewMean+fixedResults[startIndex-2].NewMean+fixedResults[startIndex-3].NewMean) / 3.0
//		} else if startIndex-2 >= 0 {
//			left3Mean = float64(fixedResults[startIndex-1].NewMean+fixedResults[startIndex-2].NewMean) / 2.0
//		} else {
//			return false, 0, 0
//		}
//		// 向右，三个或者三个位置
//		if startIndex+3 >= 0 {
//			right3Mean = float64(fixedResults[startIndex+1].NewMean+fixedResults[startIndex+2].NewMean+fixedResults[startIndex+3].NewMean) / 3.0
//		} else if startIndex+2 >= 0 {
//			right3Mean = float64(fixedResults[startIndex+1].NewMean+fixedResults[startIndex+2].NewMean) / 2.0
//		} else {
//			return false, 0, 0
//		}
//		// 将这个匹配的段中的子分段的时间轴偏移都进行一次计算，推算出到底是怎么样的配比可以得到这样的偏移结论
//		for i, info := range fixedResults[startIndex].MatchWindowInfos {
//
//			perPartLen := info.EndVADIndex - info.StartVADIndex
//			op := OverParts{}
//			// xLen 计算公式见推到公式截图
//			xLen := (info.TimeDiffStartCorrelation*float64(perPartLen) - right3Mean*float64(perPartLen)) / (left3Mean - right3Mean)
//			yLen := float64(perPartLen) - xLen
//
//			op.XLen = xLen
//			op.YLen = yLen
//			op.XMean = left3Mean
//			op.YMean = right3Mean
//
//			fixedResults[startIndex].IsOverParts = true
//			fixedResults[startIndex].MatchWindowInfos[i].OP = op
//		}
//
//		return true, fixedResults[startIndex+1].NewMean, fixedResults[startIndex+1].NewSD
//	}
//
//	return false, 0, 0
//}
//
//// slidingWindowProcessorV2 滑动窗口计算时间轴偏移
//func (s *SubTimelineFixer) slidingWindowProcessorV2(windowInfo *WindowInfo) (*MatchInfo, error) {
//
//	// -------------------------------------------------
//	var bUseSubOrAudioAsBase = true
//	if windowInfo.BaseUnit == nil && windowInfo.BaseAudioFloatList != nil {
//		// 使用 音频 来进行匹配
//		bUseSubOrAudioAsBase = false
//	} else if windowInfo.BaseUnit != nil {
//		// 使用 字幕 来进行匹配
//		bUseSubOrAudioAsBase = true
//	} else {
//		return nil, errors.New("GetOffsetTimeV2 input baseUnit or AudioVad is nil")
//	}
//	// -------------------------------------------------
//	outMatchInfo := MatchInfo{
//		IndexMatchWindowInfoMap: make(map[int]MatchWindowInfo, 0),
//		StartDiffTimeList:       make([]float64, 0),
//		StartDiffTimeMap:        treemap.NewWith(utils.Float64Comparator),
//		StartDiffTimeListEx:     make(stat.Float64Slice, 0),
//	}
//	fixFunc := func(i interface{}) error {
//		inData := i.(InputData)
//		// -------------------------------------------------
//		// 开始匹配
//		// 这里的对白单元，当前的 Base 进行对比，详细示例见图解。Step 2 中橙色的区域
//		fffAligner := NewFFTAligner(DefaultMaxOffsetSeconds, SampleRate)
//		var bok = false
//		var nowBaseStartTime = 0.0
//		var offsetIndex = 0
//		var score = 0.0
//		srcMaxLen := 0
//		// 图解，参考 Step 3
//		if bUseSubOrAudioAsBase == false {
//			// 使用 音频 来进行匹配
//			// 去掉头和尾，具体百分之多少，见 V2_FrontAndEndPerBase
//			audioCutLen := int(float64(len(inData.BaseAudioVADList)) * s.FixerConfig.V2_FrontAndEndPerBase)
//
//			srcMaxLen = windowInfo.SrcWindowLen + inData.OffsetIndex
//			if srcMaxLen >= len(inData.SrcUnit.GetVADFloatSlice()) {
//				srcMaxLen = len(inData.SrcUnit.GetVADFloatSlice()) - 1
//			}
//			offsetIndex, score = fffAligner.Fit(inData.BaseAudioVADList[audioCutLen:len(inData.BaseAudioVADList)-audioCutLen], inData.SrcUnit.GetVADFloatSlice()[inData.OffsetIndex:srcMaxLen])
//			realOffsetIndex := offsetIndex + audioCutLen
//			if realOffsetIndex < 0 {
//				return nil
//			}
//			// offsetIndex 这里得到的是 10ms 为一个单位的 OffsetIndex
//			nowBaseStartTime = vad.GetAudioIndex2Time(realOffsetIndex)
//
//		} else {
//			// 使用 字幕 来进行匹配
//
//			srcMaxLen = inData.OffsetIndex + windowInfo.SrcWindowLen
//			if srcMaxLen >= len(inData.SrcUnit.GetVADFloatSlice()) {
//				srcMaxLen = len(inData.SrcUnit.GetVADFloatSlice()) - 1
//			}
//			offsetIndex, score = fffAligner.Fit(inData.BaseUnit.GetVADFloatSlice(), inData.SrcUnit.GetVADFloatSlice()[inData.OffsetIndex:srcMaxLen])
//			if offsetIndex < 0 {
//				return nil
//			}
//			bok, nowBaseStartTime = inData.BaseUnit.GetIndexTimeNumber(offsetIndex, true)
//			if bok == false {
//				return nil
//			}
//		}
//		// 需要校正的字幕
//		bok, nowSrcStartTime := inData.SrcUnit.GetIndexTimeNumber(inData.OffsetIndex, true)
//		if bok == false {
//			return nil
//		}
//		// 时间差值
//		TimeDiffStartCorrelation := nowBaseStartTime - nowSrcStartTime
//		s.log.Debugln("------------")
//		s.log.Debugln("OffsetTime:", fmt.Sprintf("%v", TimeDiffStartCorrelation),
//			"offsetIndex:", offsetIndex,
//			"score:", fmt.Sprintf("%v", score))
//
//		mutexFixV2.Lock()
//		// 这里的未必的顺序的，所以才有 IndexMatchWindowInfoMap 的存在的意义
//		outMatchInfo.IndexMatchWindowInfoMap[inData.Index] = MatchWindowInfo{TimeDiffStartCorrelation: TimeDiffStartCorrelation,
//			StartVADIndex: inData.OffsetIndex,
//			EndVADIndex:   srcMaxLen}
//		outMatchInfo.StartDiffTimeList = append(outMatchInfo.StartDiffTimeList, TimeDiffStartCorrelation)
//		outMatchInfo.StartDiffTimeListEx = append(outMatchInfo.StartDiffTimeListEx, TimeDiffStartCorrelation)
//		outMatchInfo.StartDiffTimeMap.Put(score, windowInfo.MatchedTimes)
//		windowInfo.MatchedTimes++
//		mutexFixV2.Unlock()
//		// -------------------------------------------------
//		return nil
//	}
//	// -------------------------------------------------
//	antPool, err := ants.NewPoolWithFunc(s.FixerConfig.V2_FixThreads, func(inData interface{}) {
//		data := inData.(InputData)
//		defer data.Wg.Done()
//		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.FixerConfig.V2_SubOneUnitProcessTimeOut)*time.Second)
//		defer cancel()
//
//		done := make(chan error, 1)
//		panicChan := make(chan interface{}, 1)
//		go func() {
//			defer func() {
//				if p := recover(); p != nil {
//					panicChan <- p
//				}
//close(done)
//close(panicChan)
//			}()
//
//			done <- fixFunc(inData)
//		}()
//
//		select {
//		case err := <-done:
//			if err != nil {
//				s.log.Errorln("GetOffsetTimeV2.NewPoolWithFunc done with Error", err.Error())
//			}
//			return
//		case p := <-panicChan:
//			s.log.Errorln("GetOffsetTimeV2.NewPoolWithFunc got panic", p)
//			return
//		case <-ctx.Done():
//			s.log.Errorln("GetOffsetTimeV2.NewPoolWithFunc got time out", ctx.Err())
//			return
//		}
//	})
//	if err != nil {
//		return nil, err
//	}
//	defer antPool.Release()
//	// -------------------------------------------------
//	wg := sync.WaitGroup{}
//	index := 0
//	for i := windowInfo.SrcSlideStartIndex; i < windowInfo.SrcSlideStartIndex+windowInfo.SrcSlideLen-1; {
//		wg.Add(1)
//
//		if bUseSubOrAudioAsBase == true {
//			// 使用字幕
//			err = antPool.Invoke(InputData{Index: index, BaseUnit: *windowInfo.BaseUnit, SrcUnit: *windowInfo.SrcUnit, OffsetIndex: i, Wg: &wg})
//		} else {
//			// 使用音频
//			err = antPool.Invoke(InputData{Index: index, BaseAudioVADList: windowInfo.BaseAudioFloatList, SrcUnit: *windowInfo.SrcUnit, OffsetIndex: i, Wg: &wg})
//		}
//
//		if err != nil {
//			s.log.Errorln("GetOffsetTimeV2 ants.Invoke", err)
//		}
//
//		i += windowInfo.OneStep
//		index++
//	}
//	wg.Wait()
//
//	return &outMatchInfo, nil
//}
//
//func (s *SubTimelineFixer) calcMeanAndSDV2(startDiffTimeList stat.Float64Slice, tmpStartDiffTime []float64) FixResult {
//
//	oldMean := stat.Mean(startDiffTimeList)
//	oldSd := stat.Sd(startDiffTimeList)
//	newMean := MinValue
//	newSd := MinValue
//	per := 1.0
//
//	if len(tmpStartDiffTime) < 3 {
//		return FixResult{
//			0,
//			0,
//			oldMean,
//			oldSd,
//			oldMean,
//			oldSd,
//			per,
//			false,
//			make([]MatchWindowInfo, 0),
//		}
//	}
//
//	// 如果 SD 较大的时候才需要剔除
//	if oldSd > 0.1 {
//		var outliersMap = make(map[float64]int, 0)
//		outliers, _, _ := tukey.Outliers(0.3, tmpStartDiffTime)
//		for _, outlier := range outliers {
//			outliersMap[outlier] = 0
//		}
//		var newStartDiffTimeList = make([]float64, 0)
//		for _, f := range tmpStartDiffTime {
//
//			_, ok := outliersMap[f]
//			if ok == true {
//				continue
//			}
//
//			newStartDiffTimeList = append(newStartDiffTimeList, f)
//		}
//
//		orgLen := startDiffTimeList.Len()
//		startDiffTimeList = make(stat.Float64Slice, 0)
//		for _, f := range newStartDiffTimeList {
//			startDiffTimeList = append(startDiffTimeList, f)
//		}
//		newLen := startDiffTimeList.Len()
//
//		per = float64(newLen) / float64(orgLen)
//
//		newMean = stat.Mean(startDiffTimeList)
//		newSd = stat.Sd(startDiffTimeList)
//	}
//
//	if my_util.IsEqual(newMean, MinValue) == true {
//		newMean = oldMean
//	}
//	if my_util.IsEqual(newSd, MinValue) == true {
//		newSd = oldSd
//	}
//	return FixResult{
//		0,
//		0,
//		oldMean,
//		oldSd,
//		newMean,
//		newSd,
//		per,
//		false,
//		make([]MatchWindowInfo, 0),
//	}
//}
//
//// GetOffsetTimeV3 使用内置的字幕校正外置的字幕时间轴
//func (s *SubTimelineFixer) GetOffsetTimeV3(infoBase, infoSrc, orgFix *subparser.FileInfo, audioVadList []vad.VADInfo) error {
//
//	// -------------------------------------------------
//	var bUseSubOrAudioAsBase = true
//	if infoBase == nil && audioVadList != nil {
//		// 使用 音频 来进行匹配
//		bUseSubOrAudioAsBase = false
//	} else if infoBase != nil {
//		// 使用 字幕 来进行匹配
//		bUseSubOrAudioAsBase = true
//	} else {
//		return errors.New("GetOffsetTimeV2 input baseUnit or AudioVad is nil")
//	}
//	// -------------------------------------------------
//	audioFloatList := vad.GetFloatSlice(audioVadList)
//	baseUnitNew, err := sub_helper.GetVADInfoFeatureFromSubNew(infoBase, 0)
//	if err != nil {
//		return err
//	}
//	srcUnitNew, err := sub_helper.GetVADInfoFeatureFromSubNew(infoSrc, 0)
//	if err != nil {
//		return err
//	}
//	/*
//		上面直接得到所有的输入源，都是完整的一个文件，字幕 or 音频
//		然后根据字幕文件每一个对白进行匹配，这里就使用 V2_FrontAndEndPerSrc 进行字幕的选择，不打算从第一句话开始
//		那么假如 一共有 100 句话，V2_FrontAndEndPerSrc 是 0.2，那么就是从 20 - 80 句话进行匹配计算
//		然后 < 20 的就继承 20 的偏移，> 80 的就继承 80 的偏移即可
//		那么现在就需要从对白中开始遍历
//	*/
//	fffAligner := NewFFTAligner(DefaultMaxOffsetSeconds, SampleRate)
//	err2, done := s.caleOne(0.1, srcUnitNew, fffAligner, baseUnitNew)
//	if done {
//		return err2
//	}
//	err2, done = s.caleOne(0.2, srcUnitNew, fffAligner, baseUnitNew)
//	if done {
//		return err2
//	}
//
//	skipSubLen := int(float64(len(infoSrc.DialoguesFilter)) * s.FixerConfig.V2_FrontAndEndPerSrc)
//
//	sort.Sort(subparser.OneDialogueByStartTime(infoSrc.DialoguesFilter))
//	sort.Sort(subparser.OneDialogueByStartTime(orgFix.DialoguesFilter))
//
//	for i := 0; i < len(infoSrc.DialoguesFilter); i++ {
//
//		// 得到的是真实的时间
//		srcOneDialogueNow := infoSrc.DialoguesFilter[i]
//		srcTimeStartNow, err := my_util.ParseTime(srcOneDialogueNow.StartTime)
//		if err != nil {
//			return err
//		}
//		orgFixOneDialogueNow := orgFix.DialoguesFilter[i]
//		orgFixTimeStartNow, err := my_util.ParseTime(orgFixOneDialogueNow.StartTime)
//		if err != nil {
//			return err
//		}
//
//		println("Index:", i, "srcTimeStartOrg:", srcTimeStartNow.Format("15:04:05.000"),
//			"src-fix-offset:", my_util.Time2SecondNumber(orgFixTimeStartNow)-my_util.Time2SecondNumber(srcTimeStartNow))
//	}
//	println("------------------")
//	for i := skipSubLen; i < len(infoSrc.DialoguesFilter)-skipSubLen-1; i++ {
//
//		var bok = false
//		var nowBaseStartTime = 0.0
//		var offsetIndex = 0
//		var score = 0.0
//		const next = 20
//		const secondRange = 45
//		// -------------------------------------------------
//		srcOneDialogueNow := infoSrc.DialoguesFilter[i]
//		iNext := i + next
//		if iNext >= len(infoSrc.DialoguesFilter)-skipSubLen-1 {
//			iNext = len(infoSrc.DialoguesFilter) - skipSubLen - 1
//		}
//		srcOneDialogueNext := infoSrc.DialoguesFilter[iNext]
//		// 得到的是真实的时间
//		srcTimeStartNow, err := my_util.ParseTime(srcOneDialogueNow.StartTime)
//		if err != nil {
//			return err
//		}
//		srcTimeEndNext, err := my_util.ParseTime(srcOneDialogueNext.EndTime)
//		if err != nil {
//			return err
//		}
//		orgFixOneDialogueNow := orgFix.DialoguesFilter[i]
//		orgFixTimeStartNow, err := my_util.ParseTime(orgFixOneDialogueNow.StartTime)
//		if err != nil {
//			return err
//		}
//		// -------------------------------------------------
//		// 需要转换为 VAD 的对应 Index，需要减去 baseTime，然后根据 10ms 进行计算
//		// -------------------------------------------------
//		// Src
//		srcStartOffsetTimeNow := srcUnitNew.RealTimeToOffsetTime(srcTimeStartNow)
//		srcStartTimeVADIndexNow := int(my_util.Time2SecondNumber(srcStartOffsetTimeNow) * 100)
//
//		srcEndOffsetTimeNext := srcUnitNew.RealTimeToOffsetTime(srcTimeEndNext)
//		srcEndTimeVADIndexNext := int(my_util.Time2SecondNumber(srcEndOffsetTimeNext) * 100)
//		// -------------------------------------------------
//		if bUseSubOrAudioAsBase == false {
//			// 使用 音频 来进行匹配
//
//		} else {
//			// 使用 字幕 来进行匹配
//			// -------------------------------------------------
//			// Base
//			baseStartOffsetTimeNow := baseUnitNew.RealTimeToOffsetTime(srcTimeStartNow).Add(-secondRange * time.Second)
//			baseStartTimeVADIndexNow := int(my_util.Time2SecondNumber(baseStartOffsetTimeNow) * 100)
//
//			baseEndOffsetTimeNext := baseUnitNew.RealTimeToOffsetTime(srcTimeEndNext).Add(secondRange * time.Second)
//			baseEndTimeVADIndexNext := int(my_util.Time2SecondNumber(baseEndOffsetTimeNext) * 100)
//			if baseEndTimeVADIndexNext >= len(baseUnitNew.VADList)-1 {
//				baseEndTimeVADIndexNext = len(baseUnitNew.VADList) - 1
//			}
//			// -------------------------------------------------
//			offsetIndex, score = fffAligner.Fit(baseUnitNew.GetVADFloatSlice()[baseStartTimeVADIndexNow:baseEndTimeVADIndexNext], srcUnitNew.GetVADFloatSlice()[srcStartTimeVADIndexNow:srcEndTimeVADIndexNext])
//			if offsetIndex < 0 {
//				//return nil
//				continue
//			}
//			bok, nowBaseStartTime = baseUnitNew.GetIndexTimeNumber(baseStartTimeVADIndexNow+offsetIndex, true)
//			if bok == false {
//				return nil
//			}
//		}
//		// 需要校正的字幕
//		bok, nowSrcStartTime := srcUnitNew.GetIndexTimeNumber(srcStartTimeVADIndexNow, true)
//		if bok == false {
//			return nil
//		}
//		// 时间差值
//		TimeDiffStartCorrelation := nowBaseStartTime - nowSrcStartTime
//
//		println("Index:", i, "srcTimeStartOrg:", srcTimeStartNow.Format("15:04:05.000"),
//			"OffsetTime:", TimeDiffStartCorrelation, "Score:", score,
//			"ChangedTime:", srcTimeStartNow.Add(time.Duration(TimeDiffStartCorrelation*1000)*time.Millisecond).Format("15:04:05.000"),
//			"OrgFixTime:", orgFixTimeStartNow.Format("15:04:05.000"),
//			"src-fix-offset:", my_util.Time2SecondNumber(orgFixTimeStartNow)-my_util.Time2SecondNumber(srcTimeStartNow))
//	}
//
//	//if baseStartTimeVADIndexNow > 3600000 {
//	//	baseStartTimeVADIndexNow = 0
//	//}
//
//	println(len(audioFloatList))
//	println(len(baseUnitNew.VADList))
//	println(len(srcUnitNew.VADList))
//
//	return nil
//}
//
//func (s *SubTimelineFixer) caleOne(cutPer float64, srcUnitNew *sub_helper.SubUnit, fffAligner *FFTAligner, baseUnitNew *sub_helper.SubUnit) (error, bool) {
//	srcVADLen := len(srcUnitNew.VADList)
//	srcCutStartIndex := int(float64(srcVADLen) * cutPer)
//	offsetIndexAll, scoreAll := fffAligner.Fit(baseUnitNew.GetVADFloatSlice(), srcUnitNew.GetVADFloatSlice()[srcCutStartIndex:srcVADLen-srcCutStartIndex])
//	bok, nowBaseStartTime := baseUnitNew.GetIndexTimeNumber(0+offsetIndexAll, true)
//	if bok == false {
//		return nil, true
//	}
//	// 需要校正的字幕
//	bok, nowSrcStartTime := srcUnitNew.GetIndexTimeNumber(srcCutStartIndex, true)
//	if bok == false {
//		return nil, true
//	}
//	// 时间差值
//	TimeDiffStartCorrelation := nowBaseStartTime - nowSrcStartTime
//	bok, srcIndexCutTime := srcUnitNew.GetIndexTime(srcCutStartIndex, true)
//	if bok == false {
//		return nil, true
//	}
//	println(srcIndexCutTime.Format("15:04:05.000"), TimeDiffStartCorrelation, scoreAll)
//	return nil, false
//}
//
//const FixMask = "-fix"
//const MinValue = -9999.0
//
//var mutexFixV2 sync.Mutex
