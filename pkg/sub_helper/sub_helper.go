package sub_helper

import (
	"errors"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/subparser"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/supplier"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/archive_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/decode"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/filter"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/language"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/regex_things"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_parser_hub"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/vad"
	"github.com/sirupsen/logrus"
)

// OrganizeDlSubFiles 需要从汇总来是网站字幕中，解压对应的压缩包中的字幕出来
func OrganizeDlSubFiles(log *logrus.Logger, tmpFolderName string, subInfos []supplier.SubInfo, isMovie bool) (map[string][]string, error) {

	// 缓存列表，整理后的字幕列表
	// SxEx - []string 字幕的路径
	var siteSubInfoDict = make(map[string][]string)
	tmpFolderFullPath, err := pkg.GetTmpFolderByName(tmpFolderName)
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
		nowFileSaveFullPath := filepath.Join(tmpFolderFullPath, GetFrontNameAndOrgName(log, &subInfos[i]))
		err = pkg.WriteFile(nowFileSaveFullPath, subInfos[i].Data)
		if err != nil {
			log.Errorln("getFrontNameAndOrgName - WriteFile", nowFileSaveFullPath, "FromWhere Name TopN", subInfos[i].FromWhere, subInfos[i].Name, subInfos[i].TopN, err)
			continue
		}
		nowExt := strings.ToLower(subInfos[i].Ext)
		epsKey := pkg.GetEpisodeKeyName(subInfos[i].Season, subInfos[i].Episode)
		_, ok := siteSubInfoDict[epsKey]
		if ok == false {
			// 不存在则实例化
			siteSubInfoDict[epsKey] = make([]string, 0)
		}
		if nowExt != ".zip" && nowExt != ".tar" && nowExt != ".rar" && nowExt != ".7z" {
			// 是否是受支持的字幕类型
			if sub_parser_hub.IsSubExtWanted(nowExt) == false {
				log.Debugln("OrganizeDlSubFiles -> IsSubExtWanted == false", "Name:", subInfos[i].Name, "FileUrl:", subInfos[i].FileUrl)
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
			err = archive_helper.UnArchiveFileEx(nowFileSaveFullPath, unzipTmpFolder)
			// 解压完成后，遍历受支持的字幕列表，加入缓存列表
			if err != nil {
				log.Errorln("archiver.UnArchive", subInfos[i].FromWhere, subInfos[i].Name, subInfos[i].TopN, err)
				continue
			}
			// 搜索这个目录下的所有符合字幕格式的文件
			subFileFullPaths, err := SearchMatchedSubFileByDir(log, unzipTmpFolder)
			if err != nil {
				log.Errorln("searchMatchedSubFile", subInfos[i].FromWhere, subInfos[i].Name, subInfos[i].TopN, err)
				continue
			}
			// 这里需要给这些下载到的文件进行改名，加是从那个网站来的前缀，后续好查找
			for _, fileFullPath := range subFileFullPaths {
				if isMovie == false {
					// 连续剧的情况
					// 从解压的文件名称推断 Season 和 Episode 信息
					_, nowSeason, nowEps, err := decode.GetSeasonAndEpisodeFromSubFileName(filepath.Base(fileFullPath))
					if err != nil {
						continue
					}
					newSubName := AddFrontName(subInfos[i], filepath.Base(fileFullPath))
					newSubNameFullPath := filepath.Join(tmpFolderFullPath, newSubName)
					// 改名
					err = os.Rename(fileFullPath, newSubNameFullPath)
					if err != nil {
						log.Errorln("os.Rename", subInfos[i].FromWhere, subInfos[i].Name, subInfos[i].TopN, err)
						continue
					}
					// 加入缓存列表
					// 根据当前字幕的信息来构建 key
					SEPKey := pkg.GetEpisodeKeyName(nowSeason, nowEps)
					_, ok = siteSubInfoDict[SEPKey]
					if ok == false {
						siteSubInfoDict[SEPKey] = make([]string, 0)
					}
					siteSubInfoDict[SEPKey] = append(siteSubInfoDict[SEPKey], newSubNameFullPath)
				} else {
					// 电影的情况
					newSubName := AddFrontName(subInfos[i], filepath.Base(fileFullPath))
					newSubNameFullPath := filepath.Join(tmpFolderFullPath, newSubName)
					// 改名
					err = os.Rename(fileFullPath, newSubNameFullPath)
					if err != nil {
						log.Errorln("os.Rename", subInfos[i].FromWhere, subInfos[i].Name, subInfos[i].TopN, err)
						continue
					}
					// 加入缓存列表
					siteSubInfoDict[epsKey] = append(siteSubInfoDict[epsKey], newSubNameFullPath)
				}

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
func GetFrontNameAndOrgName(log *logrus.Logger, info *supplier.SubInfo) string {

	infoName := ""
	fileName, err := decode.GetVideoInfoFromFileName(info.Name)
	if err != nil {
		log.Warnln("", err)
		// 替换特殊字符
		infoName = pkg.ReplaceSpecString(info.Name, "x")
	} else {
		infoName = fileName.Title + "_S" + strconv.Itoa(fileName.Season) + "E" + strconv.Itoa(fileName.Episode) + filepath.Ext(info.Name)
	}
	if len(infoName) < 1 {
		infoName = pkg.RandStringBytesMaskImprSrcSB(10) + filepath.Ext(info.Name)
	}
	info.Name = infoName

	return "[" + info.FromWhere + "]_" + strconv.FormatInt(info.TopN, 10) + "_" + infoName
}

// AddFrontName 添加文件的前缀
func AddFrontName(info supplier.SubInfo, orgName string) string {
	return "[" + info.FromWhere + "]_" + strconv.FormatInt(info.TopN, 10) + "_" + orgName
}

// SearchMatchedSubFileByDir 搜索符合后缀名的视频文件，排除 Sub_SxE0 这样的文件夹中的文件
func SearchMatchedSubFileByDir(log *logrus.Logger, dir string) ([]string, error) {
	// 这里有个梗，会出现 __MACOSX 这类文件夹，那么里面会有一样的文件，需要用文件大小排除一下，至少大于 1 kb 吧
	var fileFullPathList = make([]string, 0)
	pathSep := string(os.PathSeparator)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, curFile := range files {
		fullPath := dir + pathSep + curFile.Name()
		if pkg.IsDir(fullPath) == true {
			// 需要排除 Sub_S1E0、Sub_S2E0 这样的整季的字幕文件夹，这里仅仅是缓存，不会被加载的
			matched := regex_things.RegOneSeasonSubFolderNameMatch.FindAllStringSubmatch(curFile.Name(), -1)
			if matched != nil && len(matched) > 0 {
				continue
			}
			// 内层的错误就无视了
			oneList, _ := SearchMatchedSubFileByDir(log, fullPath)
			if oneList != nil {
				fileFullPathList = append(fileFullPathList, oneList...)
			}
		} else {
			// 这里就是文件了
			if filter.SkipFileInfo(log, curFile, fullPath) == true {
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
func SearchMatchedSubFileByOneVideo(l *logrus.Logger, oneVideoFullPath string) ([]string, error) {
	dir := filepath.Dir(oneVideoFullPath)
	fileName := filepath.Base(oneVideoFullPath)
	fileName = strings.ToLower(fileName)
	fileName = strings.ReplaceAll(fileName, filepath.Ext(fileName), "")
	pathSep := string(os.PathSeparator)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var matchedSubs = make([]string, 0)

	for _, curFile := range files {
		if curFile.IsDir() {
			continue
		}
		// 这里就是文件了
		oldPath := dir + pathSep + curFile.Name()
		if filter.SkipFileInfo(l, curFile, oldPath) == true {
			continue
		}

		// 判断的时候用小写的，后续重命名的时候用原有的名称
		nowFileName := strings.ToLower(curFile.Name())
		// 后缀名得对
		if sub_parser_hub.IsSubExtWanted(filepath.Ext(nowFileName)) == false {
			continue
		}
		// 字幕文件名应该包含 视频文件名（无后缀）
		if strings.HasPrefix(nowFileName, fileName) == false {
			continue
		}

		matchedSubs = append(matchedSubs, oldPath)
	}

	return matchedSubs, nil
}

// SearchVideoMatchSubFileAndRemoveExtMark 找到找个视频目录下相匹配的字幕，同时去除这些字幕中 .default 或者 .forced 的标记。注意这两个标记不应该同时出现，否则无法正确去除
func SearchVideoMatchSubFileAndRemoveExtMark(l *logrus.Logger, oneVideoFullPath string) error {

	dir := filepath.Dir(oneVideoFullPath)
	fileName := filepath.Base(oneVideoFullPath)
	fileName = strings.ToLower(fileName)
	fileName = strings.ReplaceAll(fileName, filepath.Ext(fileName), "")
	pathSep := string(os.PathSeparator)
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, curFile := range files {
		if curFile.IsDir() {
			continue
		} else {
			// 这里就是文件了
			oldPath := dir + pathSep + curFile.Name()
			if filter.SkipFileInfo(l, curFile, oldPath) == true {
				continue
			}
			// 判断的时候用小写的，后续重命名的时候用原有的名称
			nowFileName := strings.ToLower(curFile.Name())
			// 后缀名得对
			if sub_parser_hub.IsSubExtWanted(filepath.Ext(nowFileName)) == false {
				continue
			}
			// 字幕文件名应该包含 视频文件名（无后缀）
			if strings.HasPrefix(nowFileName, fileName) == false {
				continue
			}

			if strings.Contains(nowFileName, subparser.Sub_Ext_Mark_Default+".") == true {
				// 得包含 .default. 找个关键词
				// 去除 .default.
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

	debugFolderByName, err := pkg.GetDebugFolderByName([]string{filepath.Base(seriesDir)})
	if err != nil {
		return err
	}
	files, err := os.ReadDir(debugFolderByName)
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

			fullPath := debugFolderByName + pathSep + curFile.Name()
			err = os.RemoveAll(fullPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

/*
	只针对英文字幕进行合并分散的 DialoguesFilter
	会遇到这样的字幕，如下0
	2line-The Card Counter (2021) WEBDL-1080p.chinese(inside).ass
	它的对白一句话分了两个 dialogue 去做。这样做后续字幕时间轴校正就会遇到问题，因为只有一半，匹配占比会很低
	(每一个 Dialogue 的首字母需要分析，大写和小写的占比是多少，统计一下，正常的，和上述特殊的)
	那么，就需要额外的逻辑去对 DialoguesFilterEx 进行额外的推断
	暂时考虑的方案是，英文对白每一句的开头应该是英文大写字幕，如果是小写字幕，就应该与上语句合并，且每一句的字符长度有大于一定才触发
*/
func MergeMultiDialogue4EngSubtitle(inSubParser *subparser.FileInfo) {
	merger := NewDialogueMerger()
	for _, dialogueEx := range inSubParser.DialoguesFilterEx {
		merger.Add(dialogueEx)
	}
	inSubParser.DialoguesFilterEx = merger.Get()
}

// GetVADInfoFeatureFromSub 跟下面的 GetVADInfoFeatureFromSubNeedOffsetTimeWillInsert 函数功能一致
func GetVADInfoFeatureFromSub(fileInfo *subparser.FileInfo, frontAndEndPer float64, subUnitMaxCount int, insert bool) ([]SubUnit, error) {

	return GetVADInfoFeatureFromSubNeedOffsetTimeWillInsert(fileInfo, frontAndEndPer, subUnitMaxCount, 0, insert)
}

/*
	GetVADInfoFeatureFromSubNeedOffsetTimeWillInsert 只不过这里可以加一个每一句话固定的偏移时间
	这里的字幕要求是完整的一个字幕
	1. 抽取字幕的时间片段的时候，暂定，前 15% 和后 15% 要避开，前奏、主题曲、结尾曲
	2. 将整个字幕，抽取连续 5 句对话为一个单元，提取时间片段信息
	3. 这里抽取的是特征，也就有额外的逻辑去找这个特征（本程序内会描述为“钥匙”）
*/
func GetVADInfoFeatureFromSubNeedOffsetTimeWillInsert(fileInfo *subparser.FileInfo, SkipFrontAndEndPer float64, subUnitMaxCount int, offsetTime float64, insert bool) ([]SubUnit, error) {
	if subUnitMaxCount < 0 {
		subUnitMaxCount = 0
	}

	nowDialogue := fileInfo.Dialogues

	srcSubUnitList := make([]SubUnit, 0)
	srcSubDialogueList := make([]subparser.OneDialogue, 0)
	srcOneSubUnit := NewSubUnit()

	// 最后一个对话的结束时间
	lastDialogueExTimeEnd, err := pkg.ParseTime(nowDialogue[len(nowDialogue)-1].EndTime)
	if err != nil {
		return nil, err
	}
	// 相当于总时长
	fullDuration := pkg.Time2SecondNumber(lastDialogueExTimeEnd)
	// 最低的起始时间，因为可能需要裁剪范围
	startRangeTimeMin := fullDuration * SkipFrontAndEndPer
	endRangeTimeMax := fullDuration * (1.0 - SkipFrontAndEndPer)

	println(startRangeTimeMin)
	println(endRangeTimeMax)

	for i := 0; i < len(nowDialogue); i++ {

		oneDialogueExTimeStart, err := pkg.ParseTime(nowDialogue[i].StartTime)
		if err != nil {
			return nil, err
		}
		oneDialogueExTimeEnd, err := pkg.ParseTime(nowDialogue[i].EndTime)
		if err != nil {
			return nil, err
		}

		oneStart := pkg.Time2SecondNumber(oneDialogueExTimeStart)
		if SkipFrontAndEndPer > 0 {
			if fullDuration*SkipFrontAndEndPer > oneStart || fullDuration*(1.0-SkipFrontAndEndPer) < oneStart {
				continue
			}
		}

		if nowDialogue[i].Lines == nil || len(nowDialogue[i].Lines) == 0 {
			continue
		}
		// 如果当前的这一句话，为空，或者进过正则表达式剔除特殊字符后为空，则跳过
		if pkg.ReplaceSpecString(nowDialogue[i].Lines[0], "") == "" {
			continue
		}
		// 如果当前的这一句话，为空，或者进过正则表达式剔除特殊字符后为空，则跳过
		if pkg.ReplaceSpecString(fileInfo.GetDialogueExContent(i), "") == "" {
			continue
		}
		// 低于 5句对白，则添加
		if srcOneSubUnit.GetDialogueCount() < subUnitMaxCount {
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
			srcSubDialogueList = append(srcSubDialogueList, nowDialogue[i])

		} else {
			// 用完清空
			srcSubDialogueList = make([]subparser.OneDialogue, 0)
			// 将拼凑起来的对话组成一个单元进行存储起来
			srcSubUnitList = append(srcSubUnitList, *srcOneSubUnit)
			// 然后重置
			srcOneSubUnit = NewSubUnit()
		}
	}
	if srcOneSubUnit.GetDialogueCount() > 0 {
		srcSubUnitList = append(srcSubUnitList, *srcOneSubUnit)
	}

	return srcSubUnitList, nil
}

/*
	GetVADInfoFeatureFromSubNew 将 Sub 文件转换为 VAD List 信息
*/
func GetVADInfoFeatureFromSubNew(fileInfo *subparser.FileInfo, SkipFrontAndEndPer float64) (*SubUnit, error) {

	outSubUnits := NewSubUnit()
	if len(fileInfo.Dialogues) <= 0 {
		return nil, errors.New("GetVADInfoFeatureFromSubNew fileInfo Dialogue Length is 0")
	}
	/*
		先拼凑出完整的一个 VAD List
		因为 VAD 的窗口是 10ms，那么需要多每一句话按 10 ms 的单位进行取整
		每一句话开始、结束的时间，需要向下取整
	*/
	subStartTimeFloor := pkg.MakeFloor10msMultipleFromFloat(pkg.Time2SecondNumber(fileInfo.GetStartTime()))
	subEndTimeFloor := pkg.MakeFloor10msMultipleFromFloat(pkg.Time2SecondNumber(fileInfo.GetEndTime()))
	// 如果想要从 0 时间点开始算，那么 subStartTimeFloor 这个值就需要重置到0
	subStartTimeFloor = 0
	subFullSecondTimeFloor := subEndTimeFloor - subStartTimeFloor
	// 根据这个时长就能够得到一个完整的 VAD List，然后再通过每一句对白进行 VAD 值的调整即可，这样就能够保证
	// 相同的一个字幕因为使用 ffmpeg 导出 srt 和 ass 后的，可能存在总体时间轴不一致的问题
	// 123.450 - > 12345
	vadLen := int(subFullSecondTimeFloor*100) + 2
	subVADs := make([]vad.VADInfo, vadLen)
	subStartTimeFloor10ms := subStartTimeFloor * 100
	for i := 0; i < vadLen; i++ {
		subVADs[i] = *vad.NewVADInfoBase(false, time.Duration((subStartTimeFloor10ms+float64(i))*math.Pow10(7)))
	}
	// 计算出需要截取的片段,起始和结束
	skipLen := int(float64(vadLen) * SkipFrontAndEndPer)
	skipStartIndex := skipLen
	skipEndIndex := vadLen - skipLen
	// 现在需要从 fileInfo 的每一句对白也就对应一段连续的 VAD active = true 来进行改写，记得向下取整
	lastDialogueIndex := 0
	for _, dialogue := range fileInfo.Dialogues {

		if dialogue.Lines == nil || len(dialogue.Lines) == 0 {
			continue
		}
		// 如果当前的这一句话，为空，或者进过正则表达式剔除特殊字符后为空，则跳过
		if pkg.ReplaceSpecString(dialogue.Lines[0], "") == "" {
			continue
		}
		// 字幕的开始时间
		oneDialogueStartTime, err := pkg.ParseTime(dialogue.StartTime)
		if err != nil {
			return nil, err
		}
		// 字幕的结束时间
		oneDialogueEndTime, err := pkg.ParseTime(dialogue.EndTime)
		if err != nil {
			return nil, err
		}
		// 字幕的时长，对时间进行向下取整
		oneDialogueStartTimeFloor := pkg.MakeCeil10msMultipleFromFloat(pkg.Time2SecondNumber(oneDialogueStartTime))
		oneDialogueEndTimeFloor := pkg.MakeFloor10msMultipleFromFloat(pkg.Time2SecondNumber(oneDialogueEndTime))
		// 得到一句对白的时长
		changeVADStartIndex := int(oneDialogueStartTimeFloor * 100)
		changeVADEndIndex := int(oneDialogueEndTimeFloor * 100)
		// 不能超过 最后一句话的时常
		if changeVADStartIndex > int(subEndTimeFloor*100) {
			continue
		}
		// 也不能比起始的第一句话时间轴更低
		if changeVADStartIndex < int(subStartTimeFloor10ms) {
			continue
		}
		// 当前这句话的开始和结束信息
		changerStartIndex := changeVADStartIndex - int(subStartTimeFloor10ms)
		if changerStartIndex < 0 {
			continue
		}
		changerEndIndex := changeVADEndIndex - int(subStartTimeFloor10ms)
		if changerEndIndex < 0 {
			continue
		}
		// 如果上一个对白的最后一个 OffsetIndex 连接着当前这一句的索引的 VAD 信息 active 是 true 就设置为 false
		if lastDialogueIndex == changerStartIndex {
			for i := 1; i <= 2; i++ {
				if lastDialogueIndex-i >= 0 && subVADs[lastDialogueIndex-i].Active == true {
					subVADs[lastDialogueIndex-i].Active = false
				}
			}
		}
		// 开始根据当前这句话进行 VAD 信息的设置
		// 调整之前做好的整体 VAD 的信息，符合 VAD active = true
		if changerEndIndex >= vadLen {
			changerEndIndex = vadLen - 1
		}
		for i := changerStartIndex; i <= changerEndIndex; i++ {
			subVADs[i].Active = true
		}
		lastDialogueIndex = changerEndIndex
	}

	// 截取出来当前这一段
	tmpVADList := subVADs[skipStartIndex:skipEndIndex]
	outSubUnits.VADList = tmpVADList

	tmpStartTime := time.Time{}
	tmpStartTime = tmpStartTime.Add(tmpVADList[0].Time)
	tmpEndTime := time.Time{}
	tmpEndTime = tmpEndTime.Add(tmpVADList[len(tmpVADList)-1].Time)

	outSubUnits.SetBaseTime(tmpStartTime)
	outSubUnits.SetOffsetStartTime(tmpStartTime)
	outSubUnits.SetOffsetEndTime(tmpEndTime)

	return outSubUnits, nil
}
