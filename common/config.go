package common

type Config struct {
	UseProxy bool
	HttpProxy string
	EveryTime string		// 一轮扫描字幕下载的间隔时间
	DebugMode bool
	Threads   int			// 同时并发的线程数（准确来说在go中不是线程，是 goroutine）
	SubTypePriority  int	// 字幕下载的优先级，0 是自动，1 是 srt 优先，2 是 ass/ssa 优先
	SaveMultiSub bool
	MovieFolder string
	SeriesFolder string
	AnimeFolder string
}
