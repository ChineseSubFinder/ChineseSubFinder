package sub_timeline_fixer

import (
	"errors"
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/common"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/ass"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/srt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_parser_hub"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/grd/stat"
	"github.com/james-bowman/nlp"
	"github.com/james-bowman/nlp/measures/pairwise"
	"github.com/mndrix/tukey"
	"gonum.org/v1/gonum/mat"
	"strings"
	"time"
)

// StopWordCounter 停止词统计
func StopWordCounter(inString string, per int) []string {
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

// NewTFIDF 初始化 TF-IDF
func NewTFIDF(testCorpus []string) (*nlp.Pipeline, mat.Matrix, error) {
	newCountVectoriser := nlp.NewCountVectoriser(StopWords...)
	transformer := nlp.NewTfidfTransformer()
	// set k (the number of dimensions following truncation) to 4
	reducer := nlp.NewTruncatedSVD(4)
	lsiPipeline := nlp.NewPipeline(newCountVectoriser, transformer, reducer)
	// Transform the corpus into an LSI fitting the model to the documents in the process
	lsi, err := lsiPipeline.FitTransform(testCorpus...)
	if err != nil {
		return nil, lsi, errors.New(fmt.Sprintf("Failed to process testCorpus documents because %v", err))
	}

	return lsiPipeline, lsi, nil
}

// GetOffsetTime 暂时只支持英文的基准字幕，源字幕必须是双语中英字幕
func GetOffsetTime(baseEngSubFPath, srcSubFPath string) (time.Duration, error) {
	subParserHub := sub_parser_hub.NewSubParserHub(ass.NewParser(), srt.NewParser())
	bFind, infoBase, err := subParserHub.DetermineFileTypeFromFile(baseEngSubFPath)
	if err != nil {
		return 0, err
	}
	if bFind == false {
		return 0, nil
	}
	bFind, infoSrc, err := subParserHub.DetermineFileTypeFromFile(srcSubFPath)
	if err != nil {
		return 0, err
	}
	if bFind == false {
		return 0, nil
	}

	// 构建基准语料库，目前阶段只需要考虑是 En 的就行了
	var baseCorpus = make([]string, 0)
	for _, oneDialogueEx := range infoBase.DialoguesEx {
		baseCorpus = append(baseCorpus, oneDialogueEx.EnLine)
	}
	// 初始化
	pipLine, tfidf, err := NewTFIDF(baseCorpus)
	if err != nil {
		return 0, err
	}

	/*
		确认两个字幕间的偏移，暂定的方案是两边都连续匹配上 5 个索引，再抽取一个对话的时间进行修正计算
	*/
	maxCompareDialogue := 5
	// 基线的长度
	_, docsLength := tfidf.Dims()
	var matchIndexList = make([]MatchIndex, 0)
	sc := NewSubCompare(maxCompareDialogue)
	// 开始比较相似度，默认认为是 Ch_en 就行了
	for srcIndex, srcOneDialogueEx := range infoSrc.DialoguesEx {

		// 这里只考虑 英文 的语言
		if srcOneDialogueEx.EnLine == "" {
			continue
		}
		// run the query through the same pipeline that was fitted to the corpus and
		// to project it into the same dimensional space
		queryVector, err := pipLine.Transform(srcOneDialogueEx.EnLine)
		if err != nil {
			return 0, err
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

		if sc.Add(baseIndex, srcIndex) == false {
			sc.Clear()
			sc.Add(baseIndex, srcIndex)
		}
		if sc.Check() == false {
			continue
		}

		startBaseIndex, startSrcIndex := sc.GetStartIndex()
		matchIndexList = append(matchIndexList, MatchIndex{
			BaseNowIndex: startBaseIndex,
			SrcNowIndex:  startSrcIndex,
			Similarity:   highestSimilarity,
		})

		//println(fmt.Sprintf("Similarity: %f Base[%d] %s-%s '%s' <--> Src[%d] %s-%s '%s'",
		//	highestSimilarity,
		//	baseIndex, infoBase.DialoguesEx[baseIndex].StartTime, infoBase.DialoguesEx[baseIndex].EndTime, baseCorpus[baseIndex],
		//	srcIndex, srcOneDialogueEx.StartTime, srcOneDialogueEx.EndTime, srcOneDialogueEx.EnLine))
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
			tmpBaseIndex := matchIndexItem.BaseNowIndex + i
			tmpSrcIndex := matchIndexItem.SrcNowIndex + i

			baseTimeStart, err := time.Parse(timeFormat, infoBase.DialoguesEx[tmpBaseIndex].StartTime)
			if err != nil {
				println("baseTimeStart", err)
				continue
			}
			baseTimeEnd, err := time.Parse(timeFormat, infoBase.DialoguesEx[tmpBaseIndex].EndTime)
			if err != nil {
				println("baseTimeEnd", err)
				continue
			}
			srtTimeStart, err := time.Parse(timeFormat, infoSrc.DialoguesEx[tmpSrcIndex].StartTime)
			if err != nil {
				println("srtTimeStart", err)
				continue
			}
			srtTimeEnd, err := time.Parse(timeFormat, infoSrc.DialoguesEx[tmpSrcIndex].EndTime)
			if err != nil {
				println("srtTimeEnd", err)
				continue
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

			//println(fmt.Sprintf("Diff Start-End: %s - %s Base[%d] %s-%s '%s' <--> Src[%d] %s-%s '%s'",
			//	TimeDiffStart, TimeDiffEnd,
			//	tmpBaseIndex, infoBase.DialoguesEx[tmpBaseIndex].StartTime, infoBase.DialoguesEx[tmpBaseIndex].EndTime, infoBase.DialoguesEx[tmpBaseIndex].EnLine,
			//	tmpSrcIndex, infoSrc.DialoguesEx[tmpSrcIndex].StartTime, infoSrc.DialoguesEx[tmpSrcIndex].EndTime, infoSrc.DialoguesEx[tmpSrcIndex].EnLine))
		}
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
		outliers, _, _ := tukey.Outliers(0.1, tmpStartDiffTime)
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

	err = SaveStaticLine("bar.html", infoBase.Name, infoSrc.Name,
		per, oldMean, oldSd, newMean, newSd, xAxis,
		startDiffTimeLineData, endDiffTimeLineData)
	if err != nil {
		return 0, err
	}

	return 0, nil
}

const timeFormatAss = "15:04:05.00"
const timeFormatSrt = "15:04:05,000"
