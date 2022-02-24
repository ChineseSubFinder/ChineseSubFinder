package log_hub

type OnceLog struct {
	Index    int       `json:"index"`     // 索引,0 是最近一次，依次递增
	LogLines []OneLine `json:"log_lines"` // 日志每一行的内容
}

func NewOnceLog(index int) *OnceLog {
	return &OnceLog{
		Index:    index,
		LogLines: make([]OneLine, 0),
	}
}

type OneLine struct {
	Level   string `json:"level"`   // 日志的级别
	Date    string `json:"date"`    // 日期
	Time    string `json:"time"`    // 时间
	Content string `json:"content"` // 日志的内容
}

func NewOneLine(level, date, time, content string) *OneLine {
	return &OneLine{
		Level:   level,
		Date:    date,
		Time:    time,
		Content: content,
	}
}
