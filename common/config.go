package common

type Config struct {
	UseProxy bool
	HttpProxy string
	EveryTime string
	DebugMode bool
	SaveMultiSub bool
	FoundExistSubFileThanSkip bool
	UseUnderDocker	bool	// 是否在 docker 下使用
	MovieFolder string

}
