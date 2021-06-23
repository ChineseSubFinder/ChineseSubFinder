package model

import (
	"github.com/allanpk716/ChineseSubFinder/common"
	"github.com/go-rod/rod/lib/utils"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// OrganizeDlSubFiles 需要从汇总来是网站字幕中，解压对应的压缩包中的字幕出来
func OrganizeDlSubFiles(tmpFolderName string, subInfos []common.SupplierSubInfo) (map[string][]string, error) {

	// 缓存列表，整理后的字幕列表
	// SxEx - []string 字幕的路径
	var siteSubInfoDict = make(map[string][]string)
	tmpFolderFullPath, err := GetTmpFolder(tmpFolderName)
	if err != nil {
		return nil, err
	}

	// 把后缀名给改好
	ChangeVideoExt2SubExt(subInfos)

	// 第三方的解压库，首先不支持 io.Reader 的操作，也就是得缓存到本地硬盘再读取解压
	// 且使用 walk 会无法解压 rar，得指定具体的实例，太麻烦了，直接用通用的接口得了，就是得都缓存下来再判断
	// 基于以上两点，写了一堆啰嗦的逻辑···
	for _, subInfo := range subInfos {
		// 先存下来，保存是时候需要前缀，前缀就是从那个网站下载来的
		nowFileSaveFullPath := path.Join(tmpFolderFullPath, GetFrontNameAndOrgName(subInfo))
		err = utils.OutputFile(nowFileSaveFullPath, subInfo.Data)
		if err != nil {
			GetLogger().Errorln("getFrontNameAndOrgName - OutputFile",subInfo.FromWhere, subInfo.Name, subInfo.TopN, err)
			continue
		}
		nowExt := strings.ToLower(subInfo.Ext)
		epsKey := GetEpisodeKeyName(subInfo.Season, subInfo.Episode)
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
			unzipTmpFolder := path.Join(tmpFolderFullPath, subInfo.FromWhere)
			err = os.MkdirAll(unzipTmpFolder, os.ModePerm)
			if err != nil {
				return nil, err
			}
			err = UnArchiveFile(nowFileSaveFullPath, unzipTmpFolder)
			// 解压完成后，遍历受支持的字幕列表，加入缓存列表
			if err != nil {
				GetLogger().Errorln("archiver.UnArchive", subInfo.FromWhere, subInfo.Name, subInfo.TopN, err)
				continue
			}
			// 搜索这个目录下的所有符合字幕格式的文件
			subFileFullPaths, err := SearchMatchedSubFile(unzipTmpFolder)
			if err != nil {
				GetLogger().Errorln("searchMatchedSubFile", subInfo.FromWhere, subInfo.Name, subInfo.TopN, err)
				continue
			}
			// 这里需要给这些下载到的文件进行改名，加是从那个网站来的前缀，后续好查找
			for _, fileFullPath := range subFileFullPaths {
				newSubName := AddFrontName(subInfo, filepath.Base(fileFullPath))
				newSubNameFullPath := path.Join(tmpFolderFullPath, newSubName)
				// 改名
				err = os.Rename(fileFullPath, newSubNameFullPath)
				if err != nil {
					GetLogger().Errorln("os.Rename", subInfo.FromWhere, subInfo.Name, subInfo.TopN, err)
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
func ChangeVideoExt2SubExt(subInfos []common.SupplierSubInfo) {
	for x, info := range subInfos {
		tmpSubFileName := info.Name
		// 如果后缀名是下载字幕目标的后缀名  或者 是压缩包格式的，则跳过
		if strings.Contains(tmpSubFileName, info.Ext) == true || IsWantedArchiveExtName(tmpSubFileName) == true {

		} else {
			subInfos[x].Name = tmpSubFileName + info.Ext
		}
	}
}

// FindChineseBestSubtitle 找到合适的中文字幕，优先简体双语，简体->繁体，以及 字幕类型的优先级选择
func FindChineseBestSubtitle(subs []common.SubParserFileInfo, subTypePriority int) *common.SubParserFileInfo {

	// 先傻一点实现优先双语的，之前的写法有 bug
	for _, info := range subs {
		// 找到了中文字幕
		if HasChineseLang(info.Lang) == true {
			// 字幕的优先级 0 - 原样, 1 - srt , 2 - ass/ssa
			if subTypePriority == 1 {
				// 1 - srt
				if strings.ToLower(info.Ext) == common.SubExtSRT {
					// 优先双语
					if IsBilingualSubtitle(info.Lang) == true {
						return &info
					}
				}
			} else if subTypePriority == 2 {
				//  2 - ass/ssa
				if strings.ToLower(info.Ext) == common.SubExtASS || strings.ToLower(info.Ext) == common.SubExtSSA {
					// 优先双语
					if IsBilingualSubtitle(info.Lang) == true {
						return &info
					}
				}
			}
			// 优先双语
			if IsBilingualSubtitle(info.Lang) == true {
				return &info
			}
		}
	}
	// 然后才是 chs 和 cht
	for _, info := range subs {
		// 找到了中文字幕
		if HasChineseLang(info.Lang) == true {
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
			}
			return &info
		}
	}

	return nil
}