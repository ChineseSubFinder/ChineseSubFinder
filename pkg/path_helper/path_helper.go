package path_helper

import "strings"

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

const (
	smbPrefixBase = "smb:/"
	smbPrefixFull = "smb://"

	afpPrefixBase = "afp:/"
	afpPrefixFull = "afp://"

	nfsPrefixBase = "nfs:/"
	nfsPrefixFull = "nfs://"
)
