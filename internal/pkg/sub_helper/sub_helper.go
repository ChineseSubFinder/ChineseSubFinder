package sub_helper

import (
	"github.com/allanpk716/ChineseSubFinder/internal/common"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/archive_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/decode"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/language"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	"github.com/allanpk716/ChineseSubFinder/internal/types/subparser"
	"github.com/allanpk716/ChineseSubFinder/internal/types/supplier"
	"github.com/go-rod/rod/lib/utils"
	"io/ioutil"
	"os"
	"path"
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
		nowFileSaveFullPath := path.Join(tmpFolderFullPath, GetFrontNameAndOrgName(&subInfos[i]))
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
			if IsSubExtWanted(nowExt) == false {
				continue
			}
			// 加入缓存列表
			siteSubInfoDict[epsKey] = append(siteSubInfoDict[epsKey], nowFileSaveFullPath)
		} else {
			// 那么就是需要解压的文件了
			// 解压，给一个单独的文件夹
			unzipTmpFolder := path.Join(tmpFolderFullPath, subInfos[i].FromWhere)
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
			subFileFullPaths, err := SearchMatchedSubFile(unzipTmpFolder)
			if err != nil {
				log_helper.GetLogger().Errorln("searchMatchedSubFile", subInfos[i].FromWhere, subInfos[i].Name, subInfos[i].TopN, err)
				continue
			}
			// 这里需要给这些下载到的文件进行改名，加是从那个网站来的前缀，后续好查找
			for _, fileFullPath := range subFileFullPaths {
				newSubName := AddFrontName(subInfos[i], filepath.Base(fileFullPath))
				newSubNameFullPath := path.Join(tmpFolderFullPath, newSubName)
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

// SearchMatchedSubFile 搜索符合后缀名的视频文件，排除 Sub_SxE0 这样的文件夹中的文件
func SearchMatchedSubFile(dir string) ([]string, error) {
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
			oneList, _ := SearchMatchedSubFile(fullPath)
			if oneList != nil {
				fileFullPathList = append(fileFullPathList, oneList...)
			}
		} else {
			// 这里就是文件了
			if curFile.Size() < 1000 {
				continue
			}
			if IsSubExtWanted(filepath.Ext(curFile.Name())) == true {
				fileFullPathList = append(fileFullPathList, fullPath)
			}
		}
	}
	return fileFullPathList, nil
}

// SearchVideoMatchSubFileAndRemoveDefaultMark 找到找个视频目录下相匹配的字幕，同时去除这些字幕中 .default 的标记
func SearchVideoMatchSubFileAndRemoveDefaultMark(oneVideoFullPath string) error {

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
			if IsSubExtWanted(filepath.Ext(nowFileName)) == false {
				continue
			}
			// 字幕文件名应该包含 视频文件名（无后缀）
			if strings.Contains(nowFileName, fileName) == false {
				continue
			}
			// 得包含 .default. 找个关键词
			if strings.Contains(nowFileName, types.Emby_default+".") == false {
				continue
			}
			oldPath := dir + pathSep + curFile.Name()
			newPath := dir + pathSep + strings.ReplaceAll(curFile.Name(), types.Emby_default+".", ".")
			err = os.Rename(oldPath, newPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// IsOldVersionSubPrefixName 是否是老版本的字幕命名 .chs_en[shooter] ，符合也返回这个部分＋字幕格式后缀名 .chs_en[shooter].ass, 修改后的名称
func IsOldVersionSubPrefixName(subFileName string) (bool, string, string) {

	/*
		传入的必须是字幕格式的文件，这个就再之前判断，不要在这里再判断
		传入的文件名可能有一下几种情况
		无罪之最 - S01E01 - 重建生活.chs[shooter].ass
		无罪之最 - S01E03 - 初见端倪.zh.srt
		Loki - S01E01 - Glorious Purpose WEBDL-1080p Proper.chs_en.ass
		那么就需要先剔除，字幕的格式后缀名，然后再向后取后缀名就是 .chs[shooter] or .zh
		再判断即可
	*/
	// 无罪之最 - S01E01 - 重建生活.chs[shooter].ass -> 无罪之最 - S01E01 - 重建生活.chs[shooter]
	subTypeExt := filepath.Ext(subFileName)
	subFileNameWithOutExt := strings.ReplaceAll(subFileName, subTypeExt, "")
	// .chs[shooter]
	nowExt := filepath.Ext(subFileNameWithOutExt)
	// .chs_en[shooter].ass
	orgMixExt := nowExt + subTypeExt
	orgFileNameWithOutOrgMixExt := strings.ReplaceAll(subFileName, orgMixExt, "")
	// 这里也有两种情况，一种是单字幕 SaveMultiSub: false
	// 一种的保存了多字幕 SaveMultiSub: true
	// 先判断 单字幕
	switch nowExt {
	case types.Emby_chs:
		return true, orgMixExt, makeMixSubExtString(orgFileNameWithOutOrgMixExt, types.MatchLangChs, subTypeExt, "", true)
	case types.Emby_cht:
		return true, orgMixExt, makeMixSubExtString(orgFileNameWithOutOrgMixExt, types.MatchLangCht, subTypeExt, "", false)
	case types.Emby_chs_en:
		return true, orgMixExt, makeMixSubExtString(orgFileNameWithOutOrgMixExt, types.MatchLangChsEn, subTypeExt, "", true)
	case types.Emby_cht_en:
		return true, orgMixExt, makeMixSubExtString(orgFileNameWithOutOrgMixExt, types.MatchLangChtEn, subTypeExt, "", false)
	case types.Emby_chs_jp:
		return true, orgMixExt, makeMixSubExtString(orgFileNameWithOutOrgMixExt, types.MatchLangChsJp, subTypeExt, "", true)
	case types.Emby_cht_jp:
		return true, orgMixExt, makeMixSubExtString(orgFileNameWithOutOrgMixExt, types.MatchLangChtJp, subTypeExt, "", false)
	case types.Emby_chs_kr:
		return true, orgMixExt, makeMixSubExtString(orgFileNameWithOutOrgMixExt, types.MatchLangChsKr, subTypeExt, "", true)
	case types.Emby_cht_kr:
		return true, orgMixExt, makeMixSubExtString(orgFileNameWithOutOrgMixExt, types.MatchLangChtKr, subTypeExt, "", false)
	}
	// 再判断 多字幕情况
	spStrings := strings.Split(nowExt, "[")
	if len(spStrings) != 2 {
		return false, "", ""
	}
	// 分两段来判断是否符合标准
	// 第一段
	firstOk := true
	lang := types.MatchLangChs
	site := ""
	switch spStrings[0] {
	case types.Emby_chs:
		lang = types.MatchLangChs
	case types.Emby_cht:
		lang = types.MatchLangCht
	case types.Emby_chs_en:
		lang = types.MatchLangChsEn
	case types.Emby_cht_en:
		lang = types.MatchLangChtEn
	case types.Emby_chs_jp:
		lang = types.MatchLangChsJp
	case types.Emby_cht_jp:
		lang = types.MatchLangChtJp
	case types.Emby_chs_kr:
		lang = types.MatchLangChsKr
	case types.Emby_cht_kr:
		lang = types.MatchLangChtKr
	default:
		firstOk = false
	}
	// 第二段
	secondOk := true
	tmpSecond := strings.ReplaceAll(spStrings[1], "]", "")
	switch tmpSecond {
	case common.SubSiteZiMuKu:
		site = common.SubSiteZiMuKu
	case common.SubSiteSubHd:
		site = common.SubSiteSubHd
	case common.SubSiteShooter:
		site = common.SubSiteShooter
	case common.SubSiteXunLei:
		site = common.SubSiteXunLei
	default:
		secondOk = false
	}
	// 都要符合条件
	if firstOk == true && secondOk == true {
		return true, orgMixExt, makeMixSubExtString(orgFileNameWithOutOrgMixExt, lang, subTypeExt, site, false)
	}
	return false, "", ""
}

// GenerateMixSubName 这里会生成类似的文件名 xxxx.chinese(中英,shooter)
func GenerateMixSubName(videoFileName, subExt string, subLang types.Language, extraSubPreName string) (string, string, string) {

	videoFileNameWithOutExt := strings.ReplaceAll(filepath.Base(videoFileName),
		filepath.Ext(videoFileName), "")
	note := ""
	// extraSubPreName 那个字幕网站下载的
	if extraSubPreName != "" {
		note = "," + extraSubPreName
	}
	defaultString := ".default"
	forcedString := ".forced"

	subNewName := videoFileNameWithOutExt + ".chinese" + "(" + language.Lang2ChineseString(subLang) + note + ")" + subExt
	subNewNameWithDefault := videoFileNameWithOutExt + ".chinese" + "(" + language.Lang2ChineseString(subLang) + note + ")" + defaultString + subExt
	subNewNameWithForced := videoFileNameWithOutExt + ".chinese" + "(" + language.Lang2ChineseString(subLang) + note + ")" + forcedString + subExt

	return subNewName, subNewNameWithDefault, subNewNameWithForced
}

func makeMixSubExtString(orgFileNameWithOutExt, lang string, ext, site string, beDefault bool) string {

	tmpDefault := ""
	if beDefault == true {
		tmpDefault = types.Emby_default
	}

	if site == "" {
		return orgFileNameWithOutExt + types.Emby_chinese + "(" + lang + ")" + tmpDefault + ext
	}
	return orgFileNameWithOutExt + types.Emby_chinese + "(" + lang + "," + site + ")" + tmpDefault + ext
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

var (
	regOneSeasonSubFolderNameMatch = regexp.MustCompile(`(?m)^Sub_S\dE0`)
)
