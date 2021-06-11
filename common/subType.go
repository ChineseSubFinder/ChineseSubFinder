package common

import (
	"path/filepath"
	"strings"
)

// IsSubTypeWanted 这里匹配的字幕的格式，不包含 Ext 的 . 小数点，注意，仅仅是包含关系
func IsSubTypeWanted(subName string) bool {
	nowLowerName := strings.ToLower(subName)
	if strings.Contains(nowLowerName, SubTypeASS) ||
		strings.Contains(nowLowerName, SubTypeSSA) ||
		strings.Contains(nowLowerName, SubTypeSRT) {
		return true
	}

	return false
}

// IsSubExtWanted 输入的字幕文件名，判断后缀名是否符合期望的字幕后缀名列表
func IsSubExtWanted(subName string) bool {
	inExt := filepath.Ext(subName)
	switch strings.ToLower(inExt) {
	case SubExtSSA,SubExtASS,SubExtSRT:
		return true
	default:
		return false
	}
}

const (
	SubTypeASS = "ass"
	SubTypeSSA = "ssa"
	SubTypeSRT = "srt"

	SubExtASS = ".ass"
	SubExtSSA = ".ssa"
	SubExtSRT = ".srt"
)
