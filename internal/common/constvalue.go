package common

import "time"

const HTMLTimeOut = 60 * time.Second             // HttpClient 超时时间
const OneVideoProcessTimeOut = 20*60*time.Second // 一部电影、一个连续剧，最多的处理时间
const DownloadSubsPerSite = 1                    // 默认，每个网站下载一个字幕，允许额外传参调整
const EmbyApiGetItemsLimitMin = 50
const EmbyApiGetItemsLimitMax = 50000

const (
	DebugFolder              = "debugthings"
	TmpFolder                = "tmpthings"
	DownloadSubDuring3Months = "2160h"
	DownloadSubDuring7Days   = "168h"
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