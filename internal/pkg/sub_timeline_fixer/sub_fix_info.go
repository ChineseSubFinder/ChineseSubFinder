package sub_timeline_fixer

type SubFixInfo struct {
	FixContent string // 修复后的内容
	FileName   string // 字幕的名称，包含后缀名
}

func NewSubFixInfo(fileName, fixContent string) *SubFixInfo {
	return &SubFixInfo{
		FileName:   fileName,
		FixContent: fixContent,
	}
}
