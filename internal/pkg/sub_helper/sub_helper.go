package sub_helper

import (
	"github.com/allanpk716/ChineseSubFinder/internal/common"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/archive_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/decode"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/language"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_parser_hub"
	language2 "github.com/allanpk716/ChineseSubFinder/internal/types/language"
	"github.com/allanpk716/ChineseSubFinder/internal/types/subparser"
	"github.com/allanpk716/ChineseSubFinder/internal/types/supplier"
	"github.com/go-rod/rod/lib/utils"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// OrganizeDlSubFiles 需要从汇总来是网站字幕中，解压对应的压缩包中的字幕出来
func OrganizeDlSubFiles(tmpFolderName string, subInfos []supplier.SubInfo) (map[string][]string, error) {

	// 缓存列表，整理后的字幕列表
	// SxEx - []string 字幕的路径
	var siteSubInfoDict = make(map[string][]string)
	tmpFolderFullPath, err := pkg.GetTmpFolder(tmpFolderName)
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
		epsKey := pkg.GetEpisodeKeyName(subInfos[i].Season, subInfos[i].Episode)
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
			matched := regOneSeasonSubFolderNameMatch.FindAllStringSubmatch(curFile.Name(), -1)
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
			if strings.Contains(nowFileName, language2.Sub_Ext_Mark_Default+".") == true {
				oldPath := dir + pathSep + curFile.Name()
				newPath := dir + pathSep + strings.ReplaceAll(curFile.Name(), language2.Sub_Ext_Mark_Default+".", ".")
				err = os.Rename(oldPath, newPath)
				if err != nil {
					return err
				}
			} else if strings.Contains(nowFileName, language2.Sub_Ext_Mark_Forced+".") == true {
				// 得包含 .forced. 找个关键词
				oldPath := dir + pathSep + curFile.Name()
				newPath := dir + pathSep + strings.ReplaceAll(curFile.Name(), language2.Sub_Ext_Mark_Forced+".", ".")
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
			matched := regOneSeasonSubFolderNameMatch.FindAllStringSubmatch(curFile.Name(), -1)
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
	会遇到这样的字幕，如下
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

var (
	regOneSeasonSubFolderNameMatch = regexp.MustCompile(`(?m)^Sub_S\dE0`)
)
