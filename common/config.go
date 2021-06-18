package common

type Config struct {
	UseProxy bool
	HttpProxy string
	EveryTime string		// 一轮扫描字幕下载的间隔时间
	DebugMode bool
	Threads   int			// 同时并发的线程数（准确来说在go中不是线程，是 goroutine）
	SaveMultiSub bool
	MovieFolder string
	SeriesFolder string
	AnimeFolder string
}
