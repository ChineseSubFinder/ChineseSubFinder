package sub_helper

import (
	"github.com/allanpk716/ChineseSubFinder/internal/common"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/archive_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/decode"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/language"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/regex_things"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_parser_hub"
	"github.com/allanpk716/ChineseSubFinder/internal/types/subparser"
	"github.com/allanpk716/ChineseSubFinder/internal/types/supplier"
	"github.com/go-rod/rod/lib/utils"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// OrganizeDlSubFiles 需要从汇总来是网站字幕中，解压对应的压缩包中的字幕出来
func OrganizeDlSubFiles(tmpFolderName string, subInfos []supplier.SubInfo) (map[string][]string, error) {

	// 缓存列表，整理后的字幕列表
	// SxEx - []string 字幕的路径
	var siteSubInfoDict = make(map[string][]string)
	tmpFolderFullPath, err := my_util.GetTmpFolder(tmpFolderName)
	if err != nil {
		return nil, err
	}

	// 把后缀名给改好
	ChangeVideoExt2SubExt(subInfos)

	// 第三方的解压库，首先不支持 io.Reader 的操作，也就是得缓存到本地硬盘再读取解压
	// 且使用 walk 会无法解压 rar，得指定具体的实例，太麻烦了，直接用通用的接口得了，就是得都缓存下来再判断
	// 基于以上两点，写了一堆啰嗦的逻辑···
	for i := range subInfos {
		// 先存下来，保存是时候需要前缀，前缀就是从那个网站下载来的
		nowFileSaveFullPath := filepath.Join(tmpFolderFullPath, GetFrontNameAndOrgName(&subInfos[i]))
		err = utils.OutputFile(nowFileSaveFullPath, subInfos[i].Data)
		if err != nil {
			log_helper.GetLogger().Errorln("getFrontNameAndOrgName - OutputFile", subInfos[i].FromWhere, subInfos[i].Name, subInfos[i].TopN, err)
			continue
		}
		nowExt := strings.ToLower(subInfos[i].Ext)
		epsKey := my_util.GetEpisodeKeyName(subInfos[i].Season, subInfos[i].Episode)
		_, ok := siteSubInfoDict[epsKey]
		if ok == false {
			// 不存在则实例化
			siteSubInfoDict[epsKey] = make([]string, 0)
		}
		if nowExt != ".zip" && nowExt != ".tar" && nowExt != ".rar" && nowExt != ".7z" {
			// 是否是受支持的字幕类型
			if sub_parser_hub.IsSubExtWanted(nowExt) == false {
				continue
			}
			// 加入缓存列表
			siteSubInfoDict[epsKey] = append(siteSubInfoDict[epsKey], nowFileSaveFullPath)
		} else {
			// 那么就是需要解压的文件了
			// 解压，给一个单独的文件夹
			unzipTmpFolder := filepath.Join(tmpFolderFullPath, subInfos[i].FromWhere)
			err = os.MkdirAll(unzipTmpFolder, os.ModePerm)
			if err != nil {
				return nil, err
			}
			err = archive_helper.UnArchiveFile(nowFileSaveFullPath, unzipTmpFolder)
			// 解压完成后，遍历受支持的字幕列表，加入缓存列表
			if err != nil {
				log_helper.GetLogger().Errorln("archiver.UnArchive", subInfos[i].FromWhere, subInfos[i].Name, subInfos[i].TopN, err)
				continue
			}
			// 搜索这个目录下的所有符合字幕格式的文件
			subFileFullPaths, err := SearchMatchedSubFileByDir(unzipTmpFolder)
			if err != nil {
				log_helper.GetLogger().Errorln("searchMatchedSubFile", subInfos[i].FromWhere, subInfos[i].Name, subInfos[i].TopN, err)
				continue
			}
			// 这里需要给这些下载到的文件进行改名，加是从那个网站来的前缀，后续好查找
			for _, fileFullPath := range subFileFullPaths {
				newSubName := AddFrontName(subInfos[i], filepath.Base(fileFullPath))
				newSubNameFullPath := filepath.Join(tmpFolderFullPath, newSubName)
				// 改名
				err = os.Rename(fileFullPath, newSubNameFullPath)
				if err != nil {
					log_helper.GetLogger().Errorln("os.Rename", subInfos[i].FromWhere, subInfos[i].Name, subInfos[i].TopN, err)
					continue
				}
				// 加入缓存列表
				siteSubInfoDict[epsKey] = append(siteSubInfoDict[epsKey], newSubNameFullPath)
			}
		}
	}

	return siteSubInfoDict, nil
}

// ChangeVideoExt2SubExt 检测 Name，如果是视频的后缀名就改为字幕的后缀名
func ChangeVideoExt2SubExt(subInfos []supplier.SubInfo) {
	for x, info := range subInfos {
		tmpSubFileName := info.Name
		// 如果后缀名是下载字幕目标的后缀名  或者 是压缩包格式的，则跳过
		if strings.Contains(tmpSubFileName, info.Ext) == true || archive_helper.IsWantedArchiveExtName(tmpSubFileName) == true {

		} else {
			subInfos[x].Name = tmpSubFileName + info.Ext
		}
	}
}

// SelectChineseBestBilingualSubtitle 找到合适的双语中文字幕，简体->繁体，以及 字幕类型的优先级选择
func SelectChineseBestBilingualSubtitle(subs []subparser.FileInfo, subTypePriority int) *subparser.FileInfo {

	// 先傻一点实现优先双语的，之前的写法有 bug
	for _, info := range subs {
		// 找到了中文字幕
		if language.HasChineseLang(info.Lang) == true {
			// 字幕的优先级 0 - 原样, 1 - srt , 2 - ass/ssa
			if subTypePriority == 1 {
				// 1 - srt
				if strings.ToLower(info.Ext) == common.SubExtSRT {
					// 优先双语
					if language.IsBilingualSubtitle(info.Lang) == true {
						return &info
					}
				}
			} else if subTypePriority == 2 {
				//  2 - ass/ssa
				if strings.ToLower(info.Ext) == common.SubExtASS || strings.ToLower(info.Ext) == common.SubExtSSA {
					// 优先双语
					if language.IsBilingualSubtitle(info.Lang) == true {
						return &info
					}
				}
			} else {
				// 优先双语
				if language.IsBilingualSubtitle(info.Lang) == true {
					return &info
				}
			}
		}
	}

	return nil
}

// SelectChineseBestSubtitle 找到合适的中文字幕，简体->繁体，以及 字幕类型的优先级选择
func SelectChineseBestSubtitle(subs []subparser.FileInfo, subTypePriority int) *subparser.FileInfo {

	// 先傻一点实现优先双语的，之前的写法有 bug
	for _, info := range subs {
		// 找到了中文字幕
		if language.HasChineseLang(info.Lang) == true {
			// 字幕的优先级 0 - 原样, 1 - srt , 2 - ass/ssa
			if subTypePriority == 1 {
				// 1 - srt
				if strings.ToLower(info.Ext) == common.SubExtSRT {
					return &info
				}
			} else if subTypePriority == 2 {
				//  2 - ass/ssa
				if strings.ToLower(info.Ext) == common.SubExtASS || strings.ToLower(info.Ext) == common.SubExtSSA {
					return &info
				}
			} else {
				return &info
			}
		}
	}

	return nil
}

// GetFrontNameAndOrgName 返回的名称包含，那个网站下载的，这个网站中排名第几，文件名
func GetFrontNameAndOrgName(info *supplier.SubInfo) string {

	infoName := ""
	fileName, err := decode.GetVideoInfoFromFileName(info.Name)
	if err != nil {
		log_helper.GetLogger().Warnln("", err)
		infoName = info.Name
	} else {
		infoName = fileName.Title + "_S" + strconv.Itoa(fileName.Season) + "E" + strconv.Itoa(fileName.Episode) + filepath.Ext(info.Name)
	}
	info.Name = infoName

	return "[" + info.FromWhere + "]_" + strconv.FormatInt(info.TopN, 10) + "_" + infoName
}

// AddFrontName 添加文件的前缀
func AddFrontName(info supplier.SubInfo, orgName string) string {
	return "[" + info.FromWhere + "]_" + strconv.FormatInt(info.TopN, 10) + "_" + orgName
}

// SearchMatchedSubFileByDir 搜索符合后缀名的视频文件，排除 Sub_SxE0 这样的文件夹中的文件
func SearchMatchedSubFileByDir(dir string) ([]string, error) {
	// 这里有个梗，会出现 __MACOSX 这类文件夹，那么里面会有一样的文件，需要用文件大小排除一下，至少大于 1 kb 吧
	var fileFullPathList = make([]string, 0)
	pathSep := string(os.PathSeparator)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, curFile := range files {
		fullPath := dir + pathSep + curFile.Name()
		if curFile.IsDir() {
			// 需要排除 Sub_S1E0、Sub_S2E0 这样的整季的字幕文件夹，这里仅仅是缓存，不会被加载的
			matched := regex_things.RegOneSeasonSubFolderNameMatch.FindAllStringSubmatch(curFile.Name(), -1)
			if len(matched) > 0 {
				continue
			}
			// 内层的错误就无视了
			oneList, _ := SearchMatchedSubFileByDir(fullPath)
			if oneList != nil {
				fileFullPathList = append(fileFullPathList, oneList...)
			}
		} else {
			// 这里就是文件了
			if curFile.Size() < 1000 {
				continue
			}
			if sub_parser_hub.IsSubExtWanted(filepath.Ext(curFile.Name())) == true {
				fileFullPathList = append(fileFullPathList, fullPath)
			}
		}
	}
	return fileFullPathList, nil
}

// SearchMatchedSubFileByOneVideo 搜索这个视频当前目录下匹配的字幕
func SearchMatchedSubFileByOneVideo(oneVideoFullPath string) ([]string, error) {
	dir := filepath.Dir(oneVideoFullPath)
	fileName := filepath.Base(oneVideoFullPath)
	fileName = strings.ToLower(fileName)
	fileName = strings.ReplaceAll(fileName, filepath.Ext(fileName), "")
	pathSep := string(os.PathSeparator)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var matchedSubs = make([]string, 0)

	for _, curFile := range files {
		if curFile.IsDir() {
			continue
		}
		// 这里就是文件了
		if curFile.Size() < 1000 {
			continue
		}
		// 判断的时候用小写的，后续重命名的时候用原有的名称
		nowFileName := strings.ToLower(curFile.Name())
		// 后缀名得对
		if sub_parser_hub.IsSubExtWanted(filepath.Ext(nowFileName)) == false {
			continue
		}
		// 字幕文件名应该包含 视频文件名（无后缀）
		if strings.Contains(nowFileName, fileName) == false {
			continue
		}

		oldPath := dir + pathSep + curFile.Name()
		matchedSubs = append(matchedSubs, oldPath)
	}

	return matchedSubs, nil
}

// SearchVideoMatchSubFileAndRemoveExtMark 找到找个视频目录下相匹配的字幕，同时去除这些字幕中 .default 或者 .forced 的标记。注意这两个标记不应该同时出现，否则无法正确去除
func SearchVideoMatchSubFileAndRemoveExtMark(oneVideoFullPath string) error {

	dir := filepath.Dir(oneVideoFullPath)
	fileName := filepath.Base(oneVideoFullPath)
	fileName = strings.ToLower(fileName)
	fileName = strings.ReplaceAll(fileName, filepath.Ext(fileName), "")
	pathSep := string(os.PathSeparator)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, curFile := range files {
		if curFile.IsDir() {
			continue
		} else {
			// 这里就是文件了
			if curFile.Size() < 1000 {
				continue
			}
			// 判断的时候用小写的，后续重命名的时候用原有的名称
			nowFileName := strings.ToLower(curFile.Name())
			// 后缀名得对
			if sub_parser_hub.IsSubExtWanted(filepath.Ext(nowFileName)) == false {
				continue
			}
			// 字幕文件名应该包含 视频文件名（无后缀）
			if strings.Contains(nowFileName, fileName) == false {
				continue
			}
			// 得包含 .default. 找个关键词
			if strings.Contains(nowFileName, subparser.Sub_Ext_Mark_Default+".") == true {
				oldPath := dir + pathSep + curFile.Name()
				newPath := dir + pathSep + strings.ReplaceAll(curFile.Name(), subparser.Sub_Ext_Mark_Default+".", ".")
				err = os.Rename(oldPath, newPath)
				if err != nil {
					return err
				}
			} else if strings.Contains(nowFileName, subparser.Sub_Ext_Mark_Forced+".") == true {
				// 得包含 .forced. 找个关键词
				oldPath := dir + pathSep + curFile.Name()
				newPath := dir + pathSep + strings.ReplaceAll(curFile.Name(), subparser.Sub_Ext_Mark_Forced+".", ".")
				err = os.Rename(oldPath, newPath)
				if err != nil {
					return err
				}
			} else {
				continue
			}
		}
	}

	return nil
}

// DeleteOneSeasonSubCacheFolder 删除一个连续剧中的所有一季字幕的缓存文件夹
func DeleteOneSeasonSubCacheFolder(seriesDir string) error {

	files, err := ioutil.ReadDir(seriesDir)
	if err != nil {
		return err
	}
	pathSep := string(os.PathSeparator)
	for _, curFile := range files {
		if curFile.IsDir() == true {
			matched := regex_things.RegOneSeasonSubFolderNameMatch.FindAllStringSubmatch(curFile.Name(), -1)
			if matched == nil || len(matched) < 1 {
				continue
			}

			fullPath := seriesDir + pathSep + curFile.Name()
			err = os.RemoveAll(fullPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

/*
	只针对英文字幕进行合并分散的 Dialogues
	会遇到这样的字幕，如下0
	2line-The Card Counter (2021) WEBDL-1080p.chinese(inside).ass
	它的对白一句话分了两个 dialogue 去做。这样做后续字幕时间轴校正就会遇到问题，因为只有一半，匹配占比会很低
	(每一个 Dialogue 的首字母需要分析，大写和小写的占比是多少，统计一下，正常的，和上述特殊的)
	那么，就需要额外的逻辑去对 DialoguesEx 进行额外的推断
	暂时考虑的方案是，英文对白每一句的开头应该是英文大写字幕，如果是小写字幕，就应该与上语句合并，且每一句的字符长度有大于一定才触发
*/
func MergeMultiDialogue4EngSubtitle(inSubParser *subparser.FileInfo) {
	merger := NewDialogueMerger()
	for _, dialogueEx := range inSubParser.DialoguesEx {
		merger.Add(dialogueEx)
	}
	inSubParser.DialoguesEx = merger.Get()
}

// GetVADInfoFeatureFromSub 跟下面的 GetVADInfoFeatureFromSubNeedOffsetTimeWillInsert 函数功能一致
func GetVADInfoFeatureFromSub(infoSrc *subparser.FileInfo, FrontAndEndPer float64, SubUnitMaxCount int, insert bool, kf KeyFeatures) ([]SubUnit, error) {

	return GetVADInfoFeatureFromSubNeedOffsetTimeWillInsert(infoSrc, FrontAndEndPer, SubUnitMaxCount, 0, insert, kf)
}

/*
	GetVADInfoFeatureFromSubNeedOffsetTimeWillInsert 只不过这里可以加一个每一句话固定的偏移时间
	这里的字幕要求是完整的一个字幕
	1. 抽取字幕的时间片段的时候，暂定，前 15% 和后 15% 要避开，前奏、主题曲、结尾曲
	2. 将整个字幕，抽取连续 5 句对话为一个单元，提取时间片段信息
	3. 这里抽取的是特征，也就有额外的逻辑去找这个特征（本程序内会描述为“钥匙”）
*/
func GetVADInfoFeatureFromSubNeedOffsetTimeWillInsert(infoSrc *subparser.FileInfo, SkipFrontAndEndPer float64, SubUnitMaxCount int, offsetTime float64, insert bool, kf KeyFeatures) ([]SubUnit, error) {
	if SubUnitMaxCount < 0 {
		SubUnitMaxCount = 0
	}
	srcSubUnitList := make([]SubUnit, 0)
	srcSubDialogueList := make([]subparser.OneDialogueEx, 0)
	srcOneSubUnit := NewSubUnit()

	// srcDuration
	lastDialogueExTimeEnd, err := infoSrc.ParseTime(infoSrc.DialoguesEx[len(infoSrc.DialoguesEx)-1].EndTime)
	if err != nil {
		return nil, err
	}
	srcDuration := my_util.Time2SecendNumber(lastDialogueExTimeEnd)

	for i := 0; i < len(infoSrc.DialoguesEx); i++ {

		oneDialogueExTimeStart, err := infoSrc.ParseTime(infoSrc.DialoguesEx[i].StartTime)
		if err != nil {
			return nil, err
		}
		oneDialogueExTimeEnd, err := infoSrc.ParseTime(infoSrc.DialoguesEx[i].EndTime)
		if err != nil {
			return nil, err
		}

		oneStart := my_util.Time2SecendNumber(oneDialogueExTimeStart)

		if SkipFrontAndEndPer > 0 {
			if srcDuration*SkipFrontAndEndPer > oneStart || srcDuration*(1.0-SkipFrontAndEndPer) < oneStart {
				continue
			}
		}

		// 如果当前的这一句话，为空，或者进过正则表达式剔除特殊字符后为空，则跳过
		if my_util.ReplaceSpecString(infoSrc.GetDialogueExContent(i), "") == "" {
			continue
		}
		// 低于 5句对白，则添加
		if srcOneSubUnit.GetDialogueCount() < SubUnitMaxCount {
			// 算上偏移
			offsetTimeDuration := time.Duration(offsetTime * math.Pow10(9))
			oneDialogueExTimeStart = oneDialogueExTimeStart.Add(offsetTimeDuration)
			oneDialogueExTimeEnd = oneDialogueExTimeEnd.Add(offsetTimeDuration)
			// 如果没有偏移就是 0
			if insert == true {
				srcOneSubUnit.AddAndInsert(oneDialogueExTimeStart, oneDialogueExTimeEnd)
			} else {
				srcOneSubUnit.Add(oneDialogueExTimeStart, oneDialogueExTimeEnd)
			}
			// 这一个单元的 Dialogue 需要合并起来，才能判断是否符合“钥匙”的要求
			srcSubDialogueList = append(srcSubDialogueList, infoSrc.DialoguesEx[i])

		} else {
			// 筹够那么多句话了，需要判断一次是否符合“钥匙”的要求
			tmpNowMatchKey := IsMatchKey(srcSubDialogueList, kf)
			srcOneSubUnit.IsMatchKey = tmpNowMatchKey
			// 用完清空
			srcSubDialogueList = make([]subparser.OneDialogueEx, 0)
			// 将拼凑起来的对话组成一个单元进行存储起来
			srcSubUnitList = append(srcSubUnitList, *srcOneSubUnit)
			// 然后重置
			srcOneSubUnit = NewSubUnit()
			// TODO 这里决定了插入数据的密度，有待测试
			// i = i - SubUnitMaxCount
			/*
				确认
			*/
			if tmpNowMatchKey == false {
				// 如果没有匹配上，那么就需要步进的长度短一点
				i = i - SubUnitMaxCount
			} else {
				// 如果匹配上“钥匙”了，就直接向下找另一段
				i = i - SubUnitMaxCount/2
			}
		}
	}
	if srcOneSubUnit.GetDialogueCount() > 0 {
		srcSubUnitList = append(srcSubUnitList, *srcOneSubUnit)
	}

	return srcSubUnitList, nil
}

// IsMatchKey 是否符合“钥匙”的标准
func IsMatchKey(srcSubDialogueList []subparser.OneDialogueEx, kf KeyFeatures) bool {

	/*
		这里是设置主要依赖的还是数据源，源必须有足够的对白（暂定 50 句），才可能找到这么多信息
		这里需要匹配的“钥匙”特征，先简单实现为 (这三个需要不交叉时间段)
			1. 大坑（大于 10s 的对白间隔）至少 1 个
			2. 中坑（大于 2 且小于 5s 的对白间隔）至少 3 个
			3. 小坑（大于 1 且小于 2s 的对白间隔）至少 5 个
	*/
	dialogueIntervals := make([]float64, 0)
	tmpFileInfo := subparser.FileInfo{}
	// 现在需要进行凹坑的识别，一共由多少个，间隔多少
	for i := 0; i < len(srcSubDialogueList)-1; i++ {
		startTime, err := tmpFileInfo.ParseTime(srcSubDialogueList[i+1].StartTime)
		if err != nil {
			return false
		}
		endTime, err := tmpFileInfo.ParseTime(srcSubDialogueList[i].EndTime)
		if err != nil {
			return false
		}
		// 对话间的时间间隔
		dialogueIntervals = append(dialogueIntervals, my_util.Time2SecendNumber(startTime)-my_util.Time2SecendNumber(endTime))
	}
	// big
	for _, value := range dialogueIntervals {
		if kf.Big.Match(value) == true {
			kf.Big.NowCount++
		}
		if kf.Middle.Match(value) == true {
			kf.Middle.NowCount++
		}
		if kf.Small.Match(value) == true {
			kf.Small.NowCount++
		}
	}
	// 统计到的要 >= 目标的个数
	if kf.Big.NowCount < kf.Big.LeastCount || kf.Middle.NowCount < kf.Middle.LeastCount || kf.Small.NowCount < kf.Small.LeastCount {
		return false
	}

	return true
}
