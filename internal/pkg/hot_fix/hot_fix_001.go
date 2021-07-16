package hot_fix

/*
	本模块的目标是解决开发过程中遗留的功能缺陷需要升级的问题
	之前字幕的命名不规范，现在需要进行一次批量的替换
	ch_en[shooter] -> Chinese(中英,shooter)
*/
type HotFix001 struct {
	movieRootDir string
	seriesRootDir string
}

func (h HotFix001) GetKey() string {
	return "001"
}

func (h HotFix001) Process() error {

	// 搜索所有的字幕，找到相关的字幕进行修改


	return nil
}
