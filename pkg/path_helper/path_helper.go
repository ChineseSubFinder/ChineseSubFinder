package path_helper

import (
	"path/filepath"
	"strings"
)

func FixShareFileProtocolsPath(orgPath string) string {

	orgPath = fixSMBPath(orgPath, smbPrefixBase, smbPrefixFull)
	orgPath = fixSMBPath(orgPath, afpPrefixBase, afpPrefixFull)
	orgPath = fixSMBPath(orgPath, nfsPrefixBase, nfsPrefixFull)

	return orgPath
}

func fixSMBPath(orgPath string, prefixBase, prefixFull string) string {

	if strings.HasPrefix(orgPath, prefixBase) == true {
		// 匹配最基本的前缀
		if strings.HasPrefix(orgPath, prefixFull) == true {
			// 完全符合
			return orgPath
		} else {
			// 被转义少了一个 '/'
			return strings.ReplaceAll(orgPath, prefixBase, prefixFull)
		}
	} else {
		// 无需调整，因为没得 smb 关键词
		return orgPath
	}
}

// ChangePhysicalPathToSharePath 从物理地址转换为静态文件服务器分享地址
func ChangePhysicalPathToSharePath(physicalFullPath, pathUrlMapKey, sharePrefixPath string) string {

	dirName := strings.ReplaceAll(physicalFullPath, pathUrlMapKey, "")
	outPath := filepath.Join(sharePrefixPath, dirName)
	return outPath

	/*
		// 首先需要判断这个需要替换的部分是否是包含关系
		if strings.HasPrefix(physicalFullPath, pathUrlMapKey) == true {
			dirName := strings.ReplaceAll(physicalFullPath, pathUrlMapKey, "")
			outPath := filepath.Join(sharePrefixPath, dirName)
			return outPath
		} else {
			return ""
		}
	*/
}

const (
	smbPrefixBase = "smb:/"
	smbPrefixFull = "smb://"

	afpPrefixBase = "afp:/"
	afpPrefixFull = "afp://"

	nfsPrefixBase = "nfs:/"
	nfsPrefixFull = "nfs://"
)
