package common

type ISupplier interface {
	// TODO 这里需要考虑是什么类型的视频文件，可能是 电影、连续剧、动画，这三类应该有细节上的差异
	// 比如，连续剧，有按季下载整个季字幕包，也可能是每一季的每一集下载一个
	// 电影则可以直接用 IMDB 直接下载或者削刮后的文件名去下载
	// 动画，嗯···还没啥经验，额外粗略看来，很多坑
	GetSupplierName() string

	GetSubListFromFile(filePath string) ([]SupplierSubInfo, error)

	GetSubListFromKeyword(keyword string) ([]SupplierSubInfo, error)
}