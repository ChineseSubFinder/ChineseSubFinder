package language

import iLanguage "github.com/allanpk716/ChineseSubFinder/internal/pkg/language"

// ChangeFileCoding2UTF8 自动检测文件的编码，然后转换到 UTF-8，但是导出 bytes 的时候会把头部的 BOM 信息去除
func ChangeFileCoding2UTF8(inBytes []byte) ([]byte, error) {
	return iLanguage.ChangeFileCoding2UTF8(inBytes)
}
