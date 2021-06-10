package sub_parser

type SubFileInfo struct {
	Name	string			// 字幕的名称
	Ext		string			// 字幕的后缀名
	Dialogues []OneDialogue	// 整个字幕文件的所有对话
}

// OneDialogue 一句对话
type OneDialogue struct {
	StartTime string		// 开始时间
	EndTime string			// 结束时间
	Lines	[]string		// 台词
}