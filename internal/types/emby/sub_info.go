package emby

type SubInfo struct {
	FileName        string // 文件名称
	Content         []byte // 文件的内容
	Ext             string // 文件的后缀名
	EmbyStreamIndex int    // 在 Emby Stream 中的索引
}

func NewSubInfo(fileName, ext string, embyIndex int) *SubInfo {
	sub := SubInfo{
		FileName:        fileName,
		Ext:             ext,
		EmbyStreamIndex: embyIndex,
	}
	return &sub
}
