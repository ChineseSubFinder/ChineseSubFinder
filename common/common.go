package common

import "time"

const HTMLTimeOut = 60 * time.Second	// HttpClient 超时时间
const DownloadSubsPerSite = 1 // 默认，每个网站下载一个字幕，允许额外传参调整

const (
	DebugFolder = "debugthings"
	TmpFolder = "tmpthings"
)

const (
	SubSiteZiMuKu  = "zimuku"
	SubSiteSubHd   = "subhd"
	SubSiteShooter = "shooter"
	SubSiteXunLei  = "xunlei"
)

const (
	VideoExtMp4 = ".mp4"
	VideoExtMkv = ".mkv"
	VideoExtRmvb = ".rmvb"
	VideoExtIso = ".iso"

	SubTmpFolderName = "subtmp"
)

// 需要符合 emby 的格式要求，在后缀名前面
const (
	Emby_zh = ".zh"
	Emby_en = ".en"
	//TODO 日文 韩文 emby 字幕格式要求，瞎猜的，有需要再改（目标应该是中文字幕查找，所以···应该不需要）
	Emby_jp = ".jp"
	Emby_kr = ".kr"
)