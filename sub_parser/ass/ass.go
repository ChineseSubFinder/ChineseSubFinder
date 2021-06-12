package ass

import (
	"github.com/allanpk716/ChineseSubFinder/common"
	"github.com/allanpk716/ChineseSubFinder/model"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

type Parser struct {
}

func NewParser() *Parser {
	return &Parser{}
}

func (p Parser) GetParserName() string {
	return "ass"
}

// DetermineFileTypeFromFile 确定字幕文件的类型，是双语字幕或者某一种语言等等信息
func (p Parser) DetermineFileTypeFromFile(filePath string) (*common.SubParserFileInfo, error) {
	nowExt := filepath.Ext(filePath)
	if strings.ToLower(nowExt) != common.SubExtASS && strings.ToLower(nowExt) != common.SubExtSSA {
		return nil ,nil
	}
	fBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil ,err
	}
	inBytes, err := model.ChangeFileCoding2UTF8(fBytes)
	if err != nil {
		return nil, err
	}
	return p.DetermineFileTypeFromBytes(inBytes, nowExt)
}

// DetermineFileTypeFromBytes 确定字幕文件的类型，是双语字幕或者某一种语言等等信息
func (p Parser) DetermineFileTypeFromBytes(inBytes []byte, nowExt string) (*common.SubParserFileInfo, error){
	allString :=string(inBytes)
	// 注意，需要替换掉 \r 不然正则表达式会有问题
	allString = strings.ReplaceAll(allString, "\r", "")
	re := regexp.MustCompile(regString)
	// 找到 start end text
	matched := re.FindAllStringSubmatch(allString, -1)
	if len(matched) < 1 {
		return nil ,nil
	}
	subFileInfo := common.SubParserFileInfo{}
	subFileInfo.Ext = nowExt
	subFileInfo.Dialogues = make([]common.OneDialogue, 0)
	// 这里需要统计一共有几个 \N，以及这个数量在整体行数中的比例，这样就知道是不是双语字幕了
	countLineFeed := 0
	// 先读取一次字幕文件
	for _, oneLine := range matched {
		startTime := oneLine[1]
		endTime := oneLine[2]
		nowText := oneLine[3]
		odl := common.OneDialogue{
			StartTime: startTime,
			EndTime: endTime,
		}
		odl.Lines = make([]string, 0)
		// nowText 优先移除 \h 这个是替换空格， \h 是让两个词在一行，不换行显示
		nowText = strings.ReplaceAll(nowText, `\h` , " ")
		// nowText 这个需要先把 {} 花括号内的内容给移除
		var re = regexp.MustCompile(`(?i){.*}`)
		nowText1 := re.ReplaceAllString(nowText, "")
		nowText1 = strings.TrimRight(nowText1, "\r")
		// 然后判断是否有 \N 或者 \n
		// 直接把 \n 替换为 \N 来解析
		nowText1 = strings.ReplaceAll(nowText1, `\n` , `\N`)
		if strings.Contains(nowText1,`\N`) {
			// 有，那么就需要再次切割，一般是双语字幕
			var re2 = regexp.MustCompile(`(?i)(.*)\\N(.*)`)
			for _, matched2 := range re2.FindAllStringSubmatch(nowText1, -1) {
				for i, s := range matched2 {
					if i == 0 {continue}
					odl.Lines = append(odl.Lines, s)
				}
			}
			countLineFeed++
		} else {
			// 无，则可以直接添加
			odl.Lines = append(odl.Lines, nowText1)
		}

		subFileInfo.Dialogues = append(subFileInfo.Dialogues, odl)
	}
	// 再分析
	// 需要判断每一个 Line 是啥语言，[语言的code]次数
	var langDict map[int]int
	langDict = make(map[int]int)
	for _, dialogue := range subFileInfo.Dialogues {
		model.DetectSubLangAndStatistics(dialogue.Lines, langDict)
	}
	// 从统计出来的字典，找出 Top 1 或者 2 的出来，然后计算出是什么语言的字幕
	detectLang := model.SubLangStatistics2SubLangType(float32(countLineFeed), float32(len(matched)), langDict)
	subFileInfo.Lang = detectLang
	subFileInfo.Data = inBytes
	return &subFileInfo, nil
}

const (
	// 字幕文件对话的每一行
	regString = `Dialogue: [^,.]*[0-9]*,([1-9]?[0-9]*:[0-9]*:[0-9]*.[0-9]*),([1-9]?[0-9]*:[0-9]*:[0-9]*.[0-9]*),[^,.]*,[^,.]*,[0-9]*,[0-9]*,[0-9]*,[^,.]*,(.*)`
)
