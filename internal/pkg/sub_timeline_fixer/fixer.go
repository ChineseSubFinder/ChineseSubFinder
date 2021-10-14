package sub_timeline_fixer

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/common"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/types/sub_timeline_fiexer"
	"github.com/allanpk716/ChineseSubFinder/internal/types/subparser"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/grd/stat"
	"github.com/james-bowman/nlp/measures/pairwise"
	"github.com/mndrix/tukey"
	"gonum.org/v1/gonum/mat"
	"os"
	"strings"
	"time"
)

type SubTimelineFixer struct {
	fixerConfig sub_timeline_fiexer.SubTimelineFixerConfig
}

func NewSubTimelineFixer(fixerConfig sub_timeline_fiexer.SubTimelineFixerConfig) *SubTimelineFixer {
	return &SubTimelineFixer{
		fixerConfig: fixerConfig,
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

// GetOffsetTime 暂时只支持英文的基准字幕，源字幕必须是双语中英字幕
func (s *SubTimelineFixer) GetOffsetTime(infoBase, infoSrc *subparser.FileInfo, staticLineFileSavePath string, debugInfoFileSavePath string) (bool, float64, float64, error) {

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
		//	baseIndex, infoBase.DialoguesEx[baseIndex].StartTime, infoBase.DialoguesEx[baseIndex].EndTime, baseCorpus[baseIndex],
		//	srcIndex, srcOneDialogueEx.StartTime, srcOneDialogueEx.EndTime, srcOneDialogueEx.EnLine))

		srcIndex++
	}

	timeFormat := ""
	if infoBase.Ext == common.SubExtASS || infoBase.Ext == common.SubExtSSA {
		timeFormat = timeFormatAss
	} else {
		timeFormat = timeFormatSrt
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

			baseTimeStart, err := time.Parse(timeFormat, infoBase.DialoguesEx[tmpBaseIndex].StartTime)
			if err != nil {
				return false, 0, 0, err
			}
			baseTimeEnd, err := time.Parse(timeFormat, infoBase.DialoguesEx[tmpBaseIndex].EndTime)
			if err != nil {
				return false, 0, 0, err
			}
			srtTimeStart, err := time.Parse(timeFormat, infoSrc.DialoguesEx[tmpSrcIndex].StartTime)
			if err != nil {
				return false, 0, 0, err
			}
			srtTimeEnd, err := time.Parse(timeFormat, infoSrc.DialoguesEx[tmpSrcIndex].EndTime)
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
			//	tmpBaseIndex, infoBase.DialoguesEx[tmpBaseIndex].StartTime, infoBase.DialoguesEx[tmpBaseIndex].EndTime, infoBase.DialoguesEx[tmpBaseIndex].EnLine,
			//	tmpSrcIndex, infoSrc.DialoguesEx[tmpSrcIndex].StartTime, infoSrc.DialoguesEx[tmpSrcIndex].EndTime, infoSrc.DialoguesEx[tmpSrcIndex].EnLine))
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
		err = SaveStaticLine(staticLineFileSavePath, infoBase.Name, infoSrc.Name,
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
		tmpContent := infoSrc.Name + fmt.Sprintf("Sequence match %d dialogues (< %f%%), Skip,", s.fixerConfig.MaxCompareDialogue, s.fixerConfig.MinMatchedPercent*100) + fmt.Sprintf(" %f%% ", perMatch*100)

		debugInfos = append(debugInfos, tmpContent)

		log_helper.GetLogger().Debugln(tmpContent)
	} else {
		tmpContent := infoSrc.Name + fmt.Sprintf("Sequence match %d dialogues,", s.fixerConfig.MaxCompareDialogue) + fmt.Sprintf(" %f%% ", perMatch*100)

		debugInfos = append(debugInfos, tmpContent)

		log_helper.GetLogger().Debugln(tmpContent)
	}

	// 输出调试的匹配时间轴信息的列表
	if debugInfoFileSavePath != "" {
		err = pkg.WriteStrings2File(debugInfoFileSavePath, debugInfos)
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

// FixSubTimeline 校正时间轴
func (s *SubTimelineFixer) FixSubTimeline(infoSrc *subparser.FileInfo, inOffsetTime float64, desSaveSubFileFullPath string) (string, error) {

	/*
		从解析的实例中，正常来说是可以匹配出所有的 Dialogue 对话的 Start 和 End time 的信息
		然后找到对应的字幕的文件，进行文件内容的替换来做时间轴的校正
	*/
	// 偏移时间
	offsetTime := time.Duration(inOffsetTime*1000) * time.Millisecond
	timeFormat := ""
	if infoSrc.Ext == common.SubExtASS || infoSrc.Ext == common.SubExtSSA {
		timeFormat = timeFormatAss
	} else {
		timeFormat = timeFormatSrt
	}
	fixContent := infoSrc.Content
	for _, srcOneDialogue := range infoSrc.Dialogues {

		timeStart, err := time.Parse(timeFormat, srcOneDialogue.StartTime)
		if err != nil {
			return "", err
		}
		timeEnd, err := time.Parse(timeFormat, srcOneDialogue.EndTime)
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

const timeFormatAss = "15:04:05.00"
const timeFormatSrt = "15:04:05,000"
