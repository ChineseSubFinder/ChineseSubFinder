package task_queue

import "strings"

func MergeBucketAndKeyName(bucketName, Key string) string {
	return bucketName + splitString + Key
}

func SplitMergeName(mergeName string) (bool, string, string) {
	if strings.Contains(mergeName, splitString) == false {
		return false, "", ""
	}

	splits := strings.Split(mergeName, splitString)
	if len(splits) != 2 {
		return false, "", ""
	}

	return true, splits[0], splits[1]
}

const (
	splitString = "#"

	// 每日字幕提供者的下载字幕次数，仅仅统计次数，并不确认是那个视频的字幕下载
	BucketNamePrefixSupplierDailyDownloadCounter = "SupplierDailyDownloadCounter"

	// 今日有那些视频进行了字幕的下载
	BucketNamePrefixDailyVideoDownloadCounter = "DailyVideoDownloadCounter"
)
