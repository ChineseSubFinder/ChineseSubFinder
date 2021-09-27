package sub_timeline_fixer

import (
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/ass"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/srt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_parser_hub"
	"path"
	"sort"
	"strings"
	"time"
)

// StopWordCounter 停止词统计
func StopWordCounter(instring string, per int) []string {
	statisticTimes := make(map[string]int)
	wordsLength := strings.Fields(instring)

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
	// TODO 需要加息文件的时候 DetermineFileTypeFromFile，给出 CHLines、OtherLines 的对应在哪一个时间段
	// CHLines、OtherLines 也就是要调整这两个的输出类型
}

type StopWordsPair struct {
	Name  string
	Count int
}

type StopWordsPairList []StopWordsPair

func (a StopWordsPairList) Len() int           { return len(a) }
func (a StopWordsPairList) Less(i, j int) bool { return a[i].Count < a[j].Count }
func (a StopWordsPairList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func sortMapByValue(m map[string]int) StopWordsPairList {
	p := make(StopWordsPairList, len(m))
	i := 0
	for k, v := range m {
		p[i] = StopWordsPair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(p))
	return p
}

var StopWords = []string{"a", "about", "above", "above", "across", "after", "afterwards", "again", "against", "all", "almost", "alone", "along", "already", "also", "although", "always", "am", "among", "amongst", "amoungst", "amount", "an", "and", "another", "any", "anyhow", "anyone", "anything", "anyway", "anywhere", "are", "around", "as", "at", "back", "be", "became", "because", "become", "becomes", "becoming", "been", "before", "beforehand", "behind", "being", "below", "beside", "besides", "between", "beyond", "bill", "both", "bottom", "but", "by", "call", "can", "cannot", "cant", "co", "con", "could", "couldnt", "cry", "de", "describe", "detail", "do", "done", "down", "due", "during", "each", "eg", "eight", "either", "eleven", "else", "elsewhere", "empty", "enough", "etc", "even", "ever", "every", "everyone", "everything", "everywhere", "except", "few", "fifteen", "fify", "fill", "find", "fire", "first", "five", "for", "former", "formerly", "forty", "found", "four", "from", "front", "full", "further", "get", "give", "go", "had", "has", "hasnt", "have", "he", "hence", "her", "here", "hereafter", "hereby", "herein", "hereupon", "hers", "herself", "him", "himself", "his", "how", "however", "hundred", "ie", "if", "in", "inc", "indeed", "interest", "into", "is", "it", "its", "itself", "keep", "last", "latter", "latterly", "least", "less", "ltd", "made", "many", "may", "me", "meanwhile", "might", "mill", "mine", "more", "moreover", "most", "mostly", "move", "much", "must", "my", "myself", "name", "namely", "neither", "never", "nevertheless", "next", "nine", "no", "nobody", "none", "noone", "nor", "not", "nothing", "now", "nowhere", "of", "off", "often", "on", "once", "one", "only", "onto", "or", "other", "others", "otherwise", "our", "ours", "ourselves", "out", "over", "own", "part", "per", "perhaps", "please", "put", "rather", "re", "same", "see", "seem", "seemed", "seeming", "seems", "serious", "several", "she", "should", "show", "side", "since", "sincere", "six", "sixty", "so", "some", "somehow", "someone", "something", "sometime", "sometimes", "somewhere", "still", "such", "system", "take", "ten", "than", "that", "the", "their", "them", "themselves", "then", "thence", "there", "thereafter", "thereby", "therefore", "therein", "thereupon", "these", "they", "thickv", "thin", "third", "this", "those", "though", "three", "through", "throughout", "thru", "thus", "to", "together", "too", "top", "toward", "towards", "twelve", "twenty", "two", "un", "under", "until", "up", "upon", "us", "very", "via", "was", "we", "well", "were", "what", "whatever", "when", "whence", "whenever", "where", "whereafter", "whereas", "whereby", "wherein", "whereupon", "wherever", "whether", "which", "while", "whither", "who", "whoever", "whole", "whom", "whose", "why", "will", "with", "within", "without", "would", "yet", "you", "your", "yours", "yourself", "yourselves"}
