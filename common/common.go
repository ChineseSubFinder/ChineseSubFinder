package common

import "time"

const HTMLTimeOut = 60 * time.Second	// HttpClient 超时时间
const DownloadSubsPerSite = 1 // 默认，每个网站下载一个字幕，允许额外传参调整

const (
	DebugFolder = "debugthings"
	TmpFolder = "tmpthings"
	DownloadSubDuring30Days = "720h"
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