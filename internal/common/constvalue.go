package common

import (
	"time"
)

const HTMLTimeOut = 2 * 60 * time.Second // HttpClient 超时时间
const OneMovieProcessTimeOut = 5 * 60    // 一部电影，最多的处理时间
const OneSeriesProcessTimeOut = 30 * 60  // 一部连续剧，最多的处理时间
const ScanPlayedSubTimeOut = 30 * 60     // 扫描已经播放的字幕进行缓存的时间
const DownloadSubsPerSite = 1            // 默认，每个网站下载一个字幕，允许额外传参调整
const EmbyApiGetItemsLimitMin = 50
const EmbyApiGetItemsLimitMax = 50000

const (
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
	VideoExtMp4  = ".mp4"
	VideoExtMkv  = ".mkv"
	VideoExtRmvb = ".rmvb"
	VideoExtIso  = ".iso"
	VideoExtM2ts = ".m2ts"
)

const (
	TimeFormatPoint2 = "15:04:05.00"
	TimeFormatPoint3 = "15:04:05,000"
	TimeFormatPoint4 = "15:04:05,0000"
)

const Ignore = ".ignore"

const (
	NotifyStringTellUserWait = "If you have a lot of videos, it may take a lot of time, just wait..."
)

const QueueName = "Queue"

const StaticFileName00 = "c31ViaG5Q.txt"
