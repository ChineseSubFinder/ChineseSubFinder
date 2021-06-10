package ass

import (
	"github.com/abadojack/whatlanggo"
	"github.com/allanpk716/ChineseSubFinder/common"
	"github.com/allanpk716/ChineseSubFinder/sub_parser"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

type Parser struct {
	langOptions whatlanggo.Options 	// Whitelist
}

func NewParser() *Parser {
	p := Parser{}
	p.langOptions = whatlanggo.Options{
		Whitelist: map[whatlanggo.Lang]bool{
			whatlanggo.Cmn: true,	// 中文
			whatlanggo.Eng: true,	// 英文
			whatlanggo.Jpn: true,	// 日文
			whatlanggo.Kor: true,	// 韩文
		},
	}
	return &Parser{}
}

// DetermineFileType 确定字幕文件的类型，是双语字幕或者某一种语言等等信息
func (p Parser) DetermineFileType(filePath string) (common.Language, *sub_parser.SubFileInfo, error) {
	nowExt := filepath.Ext(filePath)
	if strings.ToLower(nowExt) != common.SubExtASS && strings.ToLower(nowExt) != common.SubExtSSA {
		return common.Unknow, nil ,nil
	}
	fBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return common.Unknow, nil ,err
	}
	re := regexp.MustCompile(regString)
	// 找到 start end text
	matched := re.FindAllStringSubmatch(string(fBytes), -1)
	if len(matched) < 1 {
		return common.Unknow, nil ,nil
	}
	subFileInfo := sub_parser.SubFileInfo{}
	subFileInfo.Ext = nowExt
	subFileInfo.Dialogues = make([]sub_parser.OneDialogue, 0)
	// 这里需要统计一共有几个 \N，以及这个数量在整体行数中的比例，这样就知道是不是双语字幕了
	countLineFeed := 0
	// 先读取一次字幕文件
	for _, oneLine := range matched {
		startTime := oneLine[1]
		endTime := oneLine[2]
		nowText := oneLine[3]
		odl := sub_parser.OneDialogue{
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
	// 是不是双语字幕，定义，超过 80% 就一定是了（不可能三语吧···）
	isDouble := false
	perLines := float32(countLineFeed) / float32(len(matched))
	if perLines > 0.8 {
		isDouble = true
	}
	println(isDouble)
	// 需要判断每一个 Line 是啥语言
	for _, dialogue := range subFileInfo.Dialogues {
		for i, line := range dialogue.Lines {
			println(line)
			info := whatlanggo.DetectWithOptions(line, p.langOptions)
			// 补是语言是 info.Lang -1
			println(i, "Language:", info.Lang, info.Lang.String())
		}
	}


	return common.Unknow, &subFileInfo ,nil
}


const (
	// 字幕文件对话的每一行
	regString = `Dialogue: [^,.]*[0-9]*,([1-9]?[0-9]*:[0-9]*:[0-9]*.[0-9]*),([1-9]?[0-9]*:[0-9]*:[0-9]*.[0-9]*),[^,.]*,[^,.]*,[0-9]*,[0-9]*,[0-9]*,[^,.]*,(.*)`
)
