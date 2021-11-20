package sub_timeline_fixer

import (
	"errors"
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/calculate_curve_correlation"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/ffmpeg_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/vad"
	"github.com/allanpk716/ChineseSubFinder/internal/types/sub_timeline_fiexer"
	"github.com/allanpk716/ChineseSubFinder/internal/types/subparser"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/emirpasic/gods/utils"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/grd/stat"
	"github.com/james-bowman/nlp/measures/pairwise"
	"github.com/mndrix/tukey"
	"gonum.org/v1/gonum/mat"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type SubTimelineFixer struct {
	fixerConfig  sub_timeline_fiexer.SubTimelineFixerConfig
	ffmpegHelper *ffmpeg_helper.FFMPEGHelper
}

func NewSubTimelineFixer(fixerConfig sub_timeline_fiexer.SubTimelineFixerConfig) *SubTimelineFixer {
	return &SubTimelineFixer{
		fixerConfig:  fixerConfig,
		ffmpegHelper: ffmpeg_helper.NewFFMPEGHelper(),
	}
}

// StopWordCounter 停止词统计
func (s *SubTimelineFixer) StopWordCounter(inString string, per int) []string {
	statisticTimes := make(map[string]int)
	wordsLength := strings.Fields(inString)

	for counts, word := range wordsLength {
		// 判断key是否存在，这个word是字符串，这个counts是统计的word的次数。
		word, ok := statisticTimes[word]
		if ok {
			word = word
			statisticTimes[wordsLength[counts]] = statisticTimes[wordsLength[counts]] + 1
		} else {
			statisticTimes[wordsLength[counts]] = 1
		}
	}

	stopWords := make([]string, 0)
	mapByValue := sortMapByValue(statisticTimes)

	breakIndex := len(mapByValue) * per / 100
	for index, wordInfo := range mapByValue {
		if index > breakIndex {
			break
		}
		stopWords = append(stopWords, wordInfo.Name)
	}

	return stopWords
}

// FixSubTimeline 校正时间轴
func (s *SubTimelineFixer) FixSubTimeline(infoSrc *subparser.FileInfo, inOffsetTime float64, desSaveSubFileFullPath string) (string, error) {

	/*
		从解析的实例中，正常来说是可以匹配出所有的 Dialogue 对话的 Start 和 End time 的信息
		然后找到对应的字幕的文件，进行文件内容的替换来做时间轴的校正
	*/
	// 偏移时间
	offsetTime := time.Duration(inOffsetTime*1000) * time.Millisecond
	fixContent := infoSrc.Content
	timeFormat := infoSrc.GetTimeFormat()
	for _, srcOneDialogue := range infoSrc.Dialogues {

		timeStart, err := infoSrc.ParseTime(srcOneDialogue.StartTime)
		if err != nil {
			return "", err
		}
		timeEnd, err := infoSrc.ParseTime(srcOneDialogue.EndTime)
		if err != nil {
			return "", err
		}

		fixTimeStart := timeStart.Add(offsetTime)
		fixTimeEnd := timeEnd.Add(offsetTime)

		fixContent = strings.ReplaceAll(fixContent, srcOneDialogue.StartTime, fixTimeStart.Format(timeFormat))
		fixContent = strings.ReplaceAll(fixContent, srcOneDialogue.EndTime, fixTimeEnd.Format(timeFormat))
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

/*
	对于 V1 版本的字幕时间轴校正来说，是有特殊的前置要求的
	1. 视频要有英文字幕
	2. 外置的字幕必须是中文的双语字幕（简英、繁英）
*/

// GetOffsetTimeV1 暂时只支持英文的基准字幕，源字幕必须是双语中英字幕
func (s *SubTimelineFixer) GetOffsetTimeV1(infoBase, infoSrc *subparser.FileInfo, staticLineFileSavePath string, debugInfoFileSavePath string) (bool, float64, float64, error) {

	var debugInfos = make([]string, 0)
	// 构建基准语料库，目前阶段只需要考虑是 En 的就行了
	var baseCorpus = make([]string, 0)
	var baseDialogueFilterMap = make(map[int]int, 0)
	/*
		这里原来的写法是所有的 base 的都放进去匹配，这样会带来一些不必要的对白
		需要剔除空白。那么就需要建立一个转换的字典
	*/
	for index, oneDialogueEx := range infoBase.DialoguesEx {
		if oneDialogueEx.EnLine == "" {
			continue
		}
		baseCorpus = append(baseCorpus, oneDialogueEx.EnLine)
		baseDialogueFilterMap[len(baseCorpus)-1] = index
	}
	// 初始化
	pipLine, tfidf, err := NewTFIDF(baseCorpus)
	if err != nil {
		return false, 0, 0, err
	}

	/*
		确认两个字幕间的偏移，暂定的方案是两边都连续匹配上 5 个索引，再抽取一个对话的时间进行修正计算
	*/
	maxCompareDialogue := s.fixerConfig.MaxCompareDialogue
	// 基线的长度
	_, docsLength := tfidf.Dims()
	var matchIndexList = make([]MatchIndex, 0)
	sc := NewSubCompare(maxCompareDialogue)
	// 开始比较相似度，默认认为是 Ch_en 就行了
	for srcIndex := 0; srcIndex < len(infoSrc.DialoguesEx); {

		srcOneDialogueEx := infoSrc.DialoguesEx[srcIndex]
		// 这里只考虑 英文 的语言
		if srcOneDialogueEx.EnLine == "" {
			srcIndex++
			continue
		}
		// run the query through the same pipeline that was fitted to the corpus and
		// to project it into the same dimensional space
		queryVector, err := pipLine.Transform(srcOneDialogueEx.EnLine)
		if err != nil {
			return false, 0, 0, err
		}
		// iterate over document feature vectors (columns) in the LSI matrix and compare
		// with the query vector for similarity.  Similarity is determined by the difference
		// between the angles of the vectors known as the cosine similarity
		highestSimilarity := -1.0
		// 匹配上的基准的索引
		var baseIndex int
		// 这里理论上需要把所有的基线遍历一次，但是，一般来说，两个字幕不可能差距在 50 行
		// 这样的好处是有助于提高搜索的性能
		// 那么就以当前的 src 的位置，向前、向后各 50 来遍历
		nowMaxScanLength := srcIndex + 50
		nowMinScanLength := srcIndex - 50
		if nowMinScanLength < 0 {
			nowMinScanLength = 0
		}
		if nowMaxScanLength > docsLength {
			nowMaxScanLength = docsLength
		}
		for i := nowMinScanLength; i < nowMaxScanLength; i++ {
			similarity := pairwise.CosineSimilarity(queryVector.(mat.ColViewer).ColView(0), tfidf.(mat.ColViewer).ColView(i))
			if similarity > highestSimilarity {
				baseIndex = i
				highestSimilarity = similarity
			}
		}

		startBaseIndex, startSrcIndex := sc.GetStartIndex()
		if sc.Add(baseIndex, srcIndex) == false {
			sc.Clear()
			srcIndex = startSrcIndex + 1
			continue
			//sc.Add(baseIndex, srcIndex)
		}
		if sc.Check() == false {
			srcIndex++
			continue
		} else {
			sc.Clear()
		}

		matchIndexList = append(matchIndexList, MatchIndex{
			BaseNowIndex: startBaseIndex,
			//BaseNowIndex: baseDialogueFilterMap[startBaseIndex],
			SrcNowIndex: startSrcIndex,
			Similarity:  highestSimilarity,
		})

		//println(fmt.Sprintf("Similarity: %f Base[%d] %s-%s '%s' <--> Src[%d] %s-%s '%s'",
		//	highestSimilarity,
		//	baseIndex, infoBase.DialoguesEx[baseIndex].relativelyStartTime, infoBase.DialoguesEx[baseIndex].relativelyEndTime, baseCorpus[baseIndex],
		//	srcIndex, srcOneDialogueEx.relativelyStartTime, srcOneDialogueEx.relativelyEndTime, srcOneDialogueEx.EnLine))

		srcIndex++
	}

	var startDiffTimeLineData = make([]opts.LineData, 0)
	var endDiffTimeLineData = make([]opts.LineData, 0)
	var tmpStartDiffTime = make([]float64, 0)
	var tmpEndDiffTime = make([]float64, 0)
	var startDiffTimeList = make(stat.Float64Slice, 0)
	var endDiffTimeList = make(stat.Float64Slice, 0)
	var xAxis = make([]string, 0)
	// 上面找出了连续匹配 maxCompareDialogue：N 次的字幕语句块
	// 求出平均时间偏移
	for mIndex, matchIndexItem := range matchIndexList {

		for i := 0; i < maxCompareDialogue; i++ {
			// 这里会统计连续的这 5 句话的时间差
			//tmpBaseIndex := matchIndexItem.BaseNowIndex + i
			tmpBaseIndex := baseDialogueFilterMap[matchIndexItem.BaseNowIndex+i]
			tmpSrcIndex := matchIndexItem.SrcNowIndex + i

			baseTimeStart, err := infoBase.ParseTime(infoBase.DialoguesEx[tmpBaseIndex].StartTime)
			if err != nil {
				return false, 0, 0, err
			}
			baseTimeEnd, err := infoBase.ParseTime(infoBase.DialoguesEx[tmpBaseIndex].EndTime)
			if err != nil {
				return false, 0, 0, err
			}
			srtTimeStart, err := infoBase.ParseTime(infoSrc.DialoguesEx[tmpSrcIndex].StartTime)
			if err != nil {
				return false, 0, 0, err
			}
			srtTimeEnd, err := infoBase.ParseTime(infoSrc.DialoguesEx[tmpSrcIndex].EndTime)
			if err != nil {
				return false, 0, 0, err
			}

			TimeDiffStart := baseTimeStart.Sub(srtTimeStart)
			TimeDiffEnd := baseTimeEnd.Sub(srtTimeEnd)

			startDiffTimeLineData = append(startDiffTimeLineData, opts.LineData{Value: TimeDiffStart.Seconds()})
			endDiffTimeLineData = append(endDiffTimeLineData, opts.LineData{Value: TimeDiffEnd.Seconds()})

			tmpStartDiffTime = append(tmpStartDiffTime, TimeDiffStart.Seconds())
			tmpEndDiffTime = append(tmpEndDiffTime, TimeDiffEnd.Seconds())

			startDiffTimeList = append(startDiffTimeList, TimeDiffStart.Seconds())
			endDiffTimeList = append(endDiffTimeList, TimeDiffEnd.Seconds())

			xAxis = append(xAxis, fmt.Sprintf("%d_%d", mIndex, i))

			debugInfos = append(debugInfos, "bs "+infoBase.DialoguesEx[tmpBaseIndex].StartTime+" <-> "+infoBase.DialoguesEx[tmpBaseIndex].EndTime)
			debugInfos = append(debugInfos, "sc "+infoSrc.DialoguesEx[tmpSrcIndex].StartTime+" <-> "+infoSrc.DialoguesEx[tmpSrcIndex].EndTime)
			debugInfos = append(debugInfos, "StartDiffTime: "+fmt.Sprintf("%f", TimeDiffStart.Seconds()))
			//println(fmt.Sprintf("Diff Start-End: %s - %s Base[%d] %s-%s '%s' <--> Src[%d] %s-%s '%s'",
			//	TimeDiffStart, TimeDiffEnd,
			//	tmpBaseIndex, infoBase.DialoguesEx[tmpBaseIndex].relativelyStartTime, infoBase.DialoguesEx[tmpBaseIndex].relativelyEndTime, infoBase.DialoguesEx[tmpBaseIndex].EnLine,
			//	tmpSrcIndex, infoSrc.DialoguesEx[tmpSrcIndex].relativelyStartTime, infoSrc.DialoguesEx[tmpSrcIndex].relativelyEndTime, infoSrc.DialoguesEx[tmpSrcIndex].EnLine))
		}
		debugInfos = append(debugInfos, "---------------------------------------------")
		//println("---------------------------------------------")
	}

	oldMean := stat.Mean(startDiffTimeList)
	oldSd := stat.Sd(startDiffTimeList)
	newMean := -1.0
	newSd := -1.0
	per := 1.0

	// 如果 SD 较大的时候才需要剔除
	if oldSd > 0.1 {
		var outliersMap = make(map[float64]int, 0)
		outliers, _, _ := tukey.Outliers(0.3, tmpStartDiffTime)
		for _, outlier := range outliers {
			outliersMap[outlier] = 0
		}
		var newStartDiffTimeList = make([]float64, 0)
		for _, f := range tmpStartDiffTime {

			_, ok := outliersMap[f]
			if ok == true {
				continue
			}

			newStartDiffTimeList = append(newStartDiffTimeList, f)
		}

		orgLen := startDiffTimeList.Len()
		startDiffTimeList = make(stat.Float64Slice, 0)
		for _, f := range newStartDiffTimeList {
			startDiffTimeList = append(startDiffTimeList, f)
		}
		newLen := startDiffTimeList.Len()

		per = float64(newLen) / float64(orgLen)

		newMean = stat.Mean(startDiffTimeList)
		newSd = stat.Sd(startDiffTimeList)
	}

	if newMean == -1.0 {
		newMean = oldMean
	}
	if newSd == -1.0 {
		newSd = oldSd
	}

	// 不为空的时候，生成调试文件
	if staticLineFileSavePath != "" {
		//staticLineFileSavePath = "bar.html"
		err = SaveStaticLineV1(staticLineFileSavePath, infoBase.Name, infoSrc.Name,
			per, oldMean, oldSd, newMean, newSd, xAxis,
			startDiffTimeLineData, endDiffTimeLineData)
		if err != nil {
			return false, 0, 0, err
		}
	}

	// 跳过的逻辑是 mean 是 0 ，那么现在如果判断有问题，缓存的调试文件继续生成，然后强制返回 0 来跳过后续的逻辑
	// 这里需要考虑，找到的连续 5 句话匹配的有多少句，占比整体所有的 Dialogue 是多少，太低也需要跳过
	matchIndexLineCount := len(matchIndexList) * maxCompareDialogue
	//perMatch := float64(matchIndexLineCount) / float64(len(infoSrc.DialoguesEx))
	perMatch := float64(matchIndexLineCount) / float64(len(baseCorpus))
	if perMatch < s.fixerConfig.MinMatchedPercent {
		tmpContent := infoSrc.Name + fmt.Sprintf(" Sequence match %d dialogues (< %f%%), Skip,", s.fixerConfig.MaxCompareDialogue, s.fixerConfig.MinMatchedPercent*100) + fmt.Sprintf(" %f%% ", perMatch*100)

		debugInfos = append(debugInfos, tmpContent)

		log_helper.GetLogger().Infoln(tmpContent)
	} else {
		tmpContent := infoSrc.Name + fmt.Sprintf(" Sequence match %d dialogues,", s.fixerConfig.MaxCompareDialogue) + fmt.Sprintf(" %f%% ", perMatch*100)

		debugInfos = append(debugInfos, tmpContent)

		log_helper.GetLogger().Infoln(tmpContent)
	}

	// 输出调试的匹配时间轴信息的列表
	if debugInfoFileSavePath != "" {
		err = my_util.WriteStrings2File(debugInfoFileSavePath, debugInfos)
		if err != nil {
			return false, 0, 0, err
		}
	}
	// 虽然有条件判断是认为有问题的，但是返回值还是要填写除去的
	if perMatch < s.fixerConfig.MinMatchedPercent {
		return false, newMean, newSd, nil
	}

	return true, newMean, newSd, nil
}

// GetOffsetTimeV2 使用内置的字幕校正外置的字幕时间轴
func (s *SubTimelineFixer) GetOffsetTimeV2(infoBase, infoSrc *subparser.FileInfo, staticLineFileSavePath string, debugInfoFileSavePath string) (bool, float64, float64, error) {

	//infoBaseSubUnitList, err := sub_helper.GetVADINfoFromSub(infoBase, 0, 10000, bInsert)
	//if err != nil {
	//	return false, 0, 0, err
	//}
	//infoBaseSubUnit := infoBaseSubUnitList[0]
	//err = infoBaseSubUnit.Save2Txt("C:\\Tmp\\base.txt")
	//if err != nil {
	//	return false, 0, 0, err
	//}
	//
	//infoSrcSubUnitList, err := sub_helper.GetVADINfoFromSub(infoSrc, 0, 10000, bInsert)
	//if err != nil {
	//	return false, 0, 0, err
	//}
	//infoSrcSubUnit := infoSrcSubUnitList[0]
	//err = infoSrcSubUnit.Save2Txt("C:\\Tmp\\src.txt")
	//if err != nil {
	//	return false, 0, 0, err
	//}

	// 需要拆分成多个 unit
	srcSubUnitList, err := sub_helper.GetVADINfoFromSub(infoSrc, FrontAndEndPer, SubUnitMaxCount, bInsert, kf)
	if err != nil {
		return false, 0, 0, err
	}
	// 时间轴差值数组
	var tmpCorrelationStartDiffTime = make([]float64, 0)
	var CorrelationStartDiffTimeList = make(stat.Float64Slice, 0)

	// 调试功能，开始针对对白单元进行匹配
	for _, srcSubUnit := range srcSubUnitList {

		if srcSubUnit.IsMatchKey == false {
			continue
		}
		// 得到当前这个单元推算出来需要提取的字幕时间轴范围，这个是 Base Sub 使用的提取段
		startTimeBaseString, subBaseLength, startTimeBaseTime, _ := srcSubUnit.GetFFMPEGCutRangeString(ExpandTimeRange)
		// 导出当前的字幕文件适合与匹配的范围的临时字幕文件
		nowTmpSubBaseFPath, errString, err := s.ffmpegHelper.ExportSubArgsByTimeRange(infoBase.FileFullPath, "base", startTimeBaseString, subBaseLength)
		if err != nil {
			log_helper.GetLogger().Errorln("ExportSubArgsByTimeRange base", errString, err)
			return false, 0, 0, err
		}

		bok, nowTmpSubBaseFileInfo, err := s.ffmpegHelper.SubParserHub.DetermineFileTypeFromFile(nowTmpSubBaseFPath)
		if err != nil {
			return false, 0, 0, err
		}
		if bok == false {
			return false, 0, 0, errors.New("DetermineFileTypeFromFile == false")
		}

		nowTmpBaseSubUnitList, err := sub_helper.GetVADINfoFromSub(nowTmpSubBaseFileInfo, 0, 10000, bInsert, nil)
		if err != nil {
			return false, 0, 0, err
		}
		nowTmpBaseSubVADUnit := nowTmpBaseSubUnitList[0]

		// -------------------------------------------------
		// 开始匹配
		correlationTM := treemap.NewWith(utils.Float64Comparator)
		for i := 0; i < len(nowTmpBaseSubVADUnit.VADList); i++ {

			// 截取的长度是以当前 srcSubUnit 基准来判断的
			// 类似滑动窗口的的功能实现
			windowStartIndex := i
			windowEndIndex := i + len(srcSubUnit.VADList)
			if windowEndIndex >= len(nowTmpBaseSubVADUnit.VADList) {
				break
			}
			// Correlation
			compareSrc := srcSubUnit.GetVADFloatSlice()
			compareBase := nowTmpBaseSubVADUnit.GetVADFloatSlice()[windowStartIndex:windowEndIndex]
			correlation := calculate_curve_correlation.CalculateCurveCorrelation(compareSrc, compareBase, len(srcSubUnit.VADList))
			correlationTM.Put(correlation, i)
			//println(fmt.Sprintf("%v %v", i, correlation))
		}
		// 找到最大的数值和索引
		tmpMaxCorrelation, tmpMaxIndex := correlationTM.Max() // tmpMaxCorrelation
		if tmpMaxCorrelation == nil || tmpMaxIndex == nil {
			continue
		}

		// CalculateCurveCorrelation 计算出来的最优解
		bok, nowCorrelationBaseIndexTime := nowTmpBaseSubVADUnit.GetIndexTimeNumber(tmpMaxIndex.(int), true)
		if bok == false {
			continue
		}
		// 相似度，1 为完全匹配
		if tmpMaxCorrelation.(float64) <= MinCorrelation {
			continue
		}
		nowSrcRealTime := srcSubUnit.GetStartTimeNumber(true)
		// 时间差值
		TimeDiffStartCorrelation := nowCorrelationBaseIndexTime + my_util.Time2SecendNumber(startTimeBaseTime) - nowSrcRealTime
		// 挑匹配时间非常合适的段落出来，这个时间需要针对调试的文件进行调整
		//if TimeDiffStartCorrelation < -6.5 || TimeDiffStartCorrelation > -6.0 {
		//	continue
		//}
		// 输出调试文件
		//b, f, f2, err2 := s.debugV2(err, nowTmpSubBaseFileInfo, srcSubUnit, errString, infoSrc, nowTmpSubBaseFPath)
		//if err2 != nil {
		//	return b, f, f2, err2
		//}

		println(fmt.Sprintf("Correlation Index:%v Corre: %v DiffTime %v", tmpMaxIndex, tmpMaxCorrelation, TimeDiffStartCorrelation))
		println("-------------------")

		tmpCorrelationStartDiffTime = append(tmpCorrelationStartDiffTime, TimeDiffStartCorrelation)
		CorrelationStartDiffTimeList = append(CorrelationStartDiffTimeList, TimeDiffStartCorrelation)
	}

	outCorrelationFixResult := s.calcMeanAndSD(CorrelationStartDiffTimeList, tmpCorrelationStartDiffTime)
	println(fmt.Sprintf("Correlation Old Mean: %v SD: %v Per: %v", outCorrelationFixResult.OldMean, outCorrelationFixResult.OldSD, outCorrelationFixResult.Per))
	println(fmt.Sprintf("Correlation New Mean: %v SD: %v Per: %v", outCorrelationFixResult.NewMean, outCorrelationFixResult.NewSD, outCorrelationFixResult.Per))

	return true, outCorrelationFixResult.NewMean, outCorrelationFixResult.NewSD, nil
}

// debugV2 V2 版本的调试信息输出
func (s *SubTimelineFixer) debugV2(err error, nowTmpSubBaseFileInfo *subparser.FileInfo, srcSubUnit sub_helper.SubUnit, errString string, infoSrc *subparser.FileInfo, nowTmpSubBaseFPath string) (bool, float64, float64, error) {
	// 这里比较特殊，因为读取的字幕文件是单独切割出来的，所以默认是有偏移的们需要使用不同的函数，把偏移算进去
	nowTmpBaseSubUnitList, err := sub_helper.GetVADINfoFromSub(nowTmpSubBaseFileInfo, 0, 10000, bInsert, nil)
	if err != nil {
		return false, 0, 0, err
	}
	nowTmpBaseSubVADUnit := nowTmpBaseSubUnitList[0]

	// 导出当前的字幕文件适合与匹配的范围的临时字幕文件，这个是 Src Sub 使用的提取段
	startTimeSrcString, subSrcLength, _, _ := srcSubUnit.GetFFMPEGCutRangeString(0)
	nowTmpSubSrcFPath, errString, err := s.ffmpegHelper.ExportSubArgsByTimeRange(infoSrc.FileFullPath, "src", startTimeSrcString, subSrcLength)
	if err != nil {
		log_helper.GetLogger().Errorln("ExportSubArgsByTimeRange src", errString, err)
		return false, 0, 0, err
	}

	var nowBaseSubTimeLineData = make([]opts.LineData, 0)
	var nowBaseSubXAxis = make([]string, 0)

	var nowSrcSubTimeLineData = make([]opts.LineData, 0)
	var nowSrcSubXAxis = make([]string, 0)

	outDir := filepath.Dir(nowTmpSubBaseFPath)

	outBaseName := filepath.Base(nowTmpSubBaseFPath)
	outSrcName := filepath.Base(nowTmpSubSrcFPath)

	outBaseNameWithOutExt := strings.ReplaceAll(outBaseName, filepath.Ext(outBaseName), "")
	outSrcNameWithOutExt := strings.ReplaceAll(outSrcName, filepath.Ext(outSrcName), "")

	srcSubVADStaticLineFullPath := filepath.Join(outDir, outSrcNameWithOutExt+"_sub_src.html")
	baseSubVADStaticLineFullPath := filepath.Join(outDir, outBaseNameWithOutExt+"_sub_base.html")
	// -------------------------------------------------
	// src 导出中间文件缓存
	for _, vadInfo := range srcSubUnit.VADList {
		nowSrcSubTimeLineData = append(nowSrcSubTimeLineData, opts.LineData{Value: vadInfo.Active})
		baseTime := srcSubUnit.GetOffsetTimeNumber()
		nowVADInfoTimeNumber := vadInfo.Time.Seconds()
		nowOffsetTime := nowVADInfoTimeNumber - baseTime
		nowSrcSubXAxis = append(nowSrcSubXAxis, fmt.Sprintf("%f", nowOffsetTime))
	}
	err = SaveStaticLineV2("Sub src", srcSubVADStaticLineFullPath, nowSrcSubXAxis, nowSrcSubTimeLineData)
	if err != nil {
		return false, 0, 0, err
	}
	// -------------------------------------------------
	// base 导出中间文件缓存
	for _, vadInfo := range nowTmpBaseSubVADUnit.VADList {
		nowBaseSubTimeLineData = append(nowBaseSubTimeLineData, opts.LineData{Value: vadInfo.Active})
		nowVADInfoTimeNumber := vadInfo.Time.Seconds()
		nowBaseSubXAxis = append(nowBaseSubXAxis, fmt.Sprintf("%f", nowVADInfoTimeNumber))
	}
	err = SaveStaticLineV2("Sub base", baseSubVADStaticLineFullPath, nowBaseSubXAxis, nowBaseSubTimeLineData)
	if err != nil {
		return false, 0, 0, err
	}
	return false, 0, 0, nil
}

func (s *SubTimelineFixer) calcMeanAndSD(startDiffTimeList stat.Float64Slice, tmpStartDiffTime []float64) FixResult {

	oldMean := stat.Mean(startDiffTimeList)
	oldSd := stat.Sd(startDiffTimeList)
	newMean := -1.0
	newSd := -1.0
	per := 1.0

	if len(tmpStartDiffTime) < 3 {
		return FixResult{
			oldMean,
			oldSd,
			newMean,
			newSd,
			per,
		}
	}

	// 如果 SD 较大的时候才需要剔除
	if oldSd > 0.1 {
		var outliersMap = make(map[float64]int, 0)
		outliers, _, _ := tukey.Outliers(0.3, tmpStartDiffTime)
		for _, outlier := range outliers {
			outliersMap[outlier] = 0
		}
		var newStartDiffTimeList = make([]float64, 0)
		for _, f := range tmpStartDiffTime {

			_, ok := outliersMap[f]
			if ok == true {
				continue
			}

			newStartDiffTimeList = append(newStartDiffTimeList, f)
		}

		orgLen := startDiffTimeList.Len()
		startDiffTimeList = make(stat.Float64Slice, 0)
		for _, f := range newStartDiffTimeList {
			startDiffTimeList = append(startDiffTimeList, f)
		}
		newLen := startDiffTimeList.Len()

		per = float64(newLen) / float64(orgLen)

		newMean = stat.Mean(startDiffTimeList)
		newSd = stat.Sd(startDiffTimeList)
	}

	if newMean == -1.0 {
		newMean = oldMean
	}
	if newSd == -1.0 {
		newSd = oldSd
	}
	return FixResult{
		oldMean,
		oldSd,
		newMean,
		newSd,
		per,
	}
}

// GetOffsetTimeV3 使用 VAD 检测语音是否有人声，输出连续的点标记，再通过 SimHash 进行匹配，找到最佳的偏移时间
func (s *SubTimelineFixer) GetOffsetTimeV3(audioInfo vad.AudioInfo, infoSrc *subparser.FileInfo, staticLineFileSavePath string, debugInfoFileSavePath string) (bool, float64, float64, error) {

	//audioVADInfos, err := vad.GetVADInfoFromAudio(vad.AudioInfo{
	//	FileFullPath: audioInfo.FileFullPath,
	//	SampleRate:   16000,
	//	BitDepth:     16,
	//})
	//if err != nil {
	//	return false, 0, 0, err
	//}
	//
	//subUnit := sub_helper.NewSubUnit()
	//subUnit.VADList = audioVADInfos
	//err = subUnit.Save2Txt("C:\\Tmp\\audio.txt")
	//if err != nil {
	//	return false, 0, 0, err
	//}
	/*
		分割字幕成若干段，然后得到若干段的时间轴，将这些段从字幕文字转换成 VADInfo
		从上面若干段时间轴，把音频给分割成多段
		然后使用 simhash 的进行比较，输出分析的曲线图等信息
	*/

	//bok, duration, err := s.ffmpegHelper.GetAudioInfo(audioInfo.FileFullPath)
	//if err != nil || bok == false {
	//	return false, 0, 0, err
	//}

	/*
		这里的字幕要求是完整的一个字幕
		1. 抽取字幕的时间片段的时候，暂定，前 15% 和后 15% 要避开，前奏、主题曲、结尾曲
		2. 将整个字幕，抽取连续 5 句对话为一个单元，提取时间片段信息
	*/
	subUnitList, err := sub_helper.GetVADINfoFromSub(infoSrc, FrontAndEndPer, SubUnitMaxCount, bInsert, nil)
	if err != nil {
		return false, 0, 0, err
	}
	// 开始针对对白单元进行匹配
	for _, subUnit := range subUnitList {

		startTimeString, subLength, _, _ := subUnit.GetFFMPEGCutRangeString(ExpandTimeRange)
		// 导出当前的音频文件适合与匹配的范围的临时音频文件
		outAudioFPath, _, errString, err := s.ffmpegHelper.ExportAudioAndSubArgsByTimeRange(audioInfo.FileFullPath, infoSrc.FileFullPath, startTimeString, subLength)
		if err != nil {
			log_helper.GetLogger().Errorln("ExportAudioAndSubArgsByTimeRange", errString, err)
			return false, 0, 0, err
		}

		audioVADInfos, err := vad.GetVADInfoFromAudio(vad.AudioInfo{
			FileFullPath: outAudioFPath,
			SampleRate:   16000,
			BitDepth:     16,
		}, false)
		if err != nil {
			return false, 0, 0, err
		}

		var subTimeLineData = make([]opts.LineData, 0)
		var subTimeLineFFTData = make([]opts.LineData, 0)
		var subXAxis = make([]string, 0)

		var audioTimeLineData = make([]opts.LineData, 0)
		var audioTimeLineFFTData = make([]opts.LineData, 0)
		var audioXAxis = make([]string, 0)

		subBuf := make([]complex128, my_util.MakePowerOfTwo(int64(len(subUnit.VADList))))
		audioBuf := make([]complex128, my_util.MakePowerOfTwo(int64(len(audioVADInfos))))
		for index, vadInfo := range subUnit.VADList {

			subTimeLineData = append(subTimeLineData, opts.LineData{Value: vadInfo.Active})
			baseTime := subUnit.GetOffsetTimeNumber()
			nowVADInfoTimeNumber := vadInfo.Time.Seconds()
			//println(fmt.Sprintf("%d - %f", index, nowVADInfoTimeNumber-baseTime))
			nowOffsetTime := nowVADInfoTimeNumber - baseTime
			subXAxis = append(subXAxis, fmt.Sprintf("%f", nowOffsetTime))

			subBuf[index] = complex(float64(my_util.Bool2Int(vadInfo.Active)), nowOffsetTime)
		}

		for i := 0; i < len(subUnit.VADList); i++ {
			subTimeLineFFTData = append(subTimeLineFFTData, opts.LineData{Value: real(subBuf[i])})
		}

		outDir := filepath.Dir(outAudioFPath)
		outBaseName := filepath.Base(outAudioFPath)
		outBaseNameWithOutExt := strings.ReplaceAll(outBaseName, filepath.Ext(outBaseName), "")

		subVADStaticLineFullPath := filepath.Join(outDir, outBaseNameWithOutExt+"_sub.html")

		err = SaveStaticLineV3("Sub", subVADStaticLineFullPath, subXAxis, subTimeLineData, subTimeLineFFTData)
		if err != nil {
			return false, 0, 0, err
		}

		for index, vadInfo := range audioVADInfos {

			audioTimeLineData = append(audioTimeLineData, opts.LineData{Value: vadInfo.Active})
			audioXAxis = append(audioXAxis, fmt.Sprintf("%f", vadInfo.Time.Seconds()))

			audioBuf[index] = complex(float64(my_util.Bool2Int(vadInfo.Active)), vadInfo.Time.Seconds())
		}

		for i := 0; i < len(audioBuf); i++ {
			audioTimeLineFFTData = append(audioTimeLineFFTData, opts.LineData{Value: real(audioBuf[i])})
		}

		audioVADStaticLineFullPath := filepath.Join(outDir, outBaseNameWithOutExt+"_audio.html")

		err = SaveStaticLineV3("Audio", audioVADStaticLineFullPath, audioXAxis, audioTimeLineData, audioTimeLineFFTData)
		if err != nil {
			return false, 0, 0, err
		}
	}

	return false, -1, -1, nil
}

const FixMask = "-fix"
const bInsert = true        // 是否插入点
const whichOne = 0          // 所有，whichOne = 1 只有 Start 的点
const FrontAndEndPer = 0.05 // 前百分之 15 和后百分之 15 都不进行识别
const SubUnitMaxCount = 30  // 一个 Sub单元有五句对白
const ExpandTimeRange = 50  // 从字幕的时间轴片段需要向前和向后多匹配一部分的音频，这里定义的就是这个 range 以分钟为单位， 正负 60 秒
const KeyPer = 0.1          // 钥匙凹坑的占比
const MinCorrelation = 0.50 // 最低的匹配度
const DTW_Radius = 1000     // DTW 半径

var kf = sub_helper.NewKeyFeatures(
	sub_helper.NewFeature(10.0, 999999, 2),
	sub_helper.NewFeature(3.0, 10.0, 5),
	sub_helper.NewFeature(1.0, 3.0, 5),
)
