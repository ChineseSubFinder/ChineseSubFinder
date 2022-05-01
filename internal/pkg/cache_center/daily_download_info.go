package cache_center

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/cache_center/models"
	"time"
)

// GetDailyDownloadCount 根据字幕提供者的名称，获取今日下载计数的次数，仅仅统计次数，并不确认是那个视频的字幕下载
// whichDay nowTime.Format("2006-01-02")
func (c *CacheCenter) GetDailyDownloadCount(supplierName string, publicIP string, whichDay ...string) (int, error) {
	defer c.locker.Unlock()
	c.locker.Lock()

	var dailyDownloadInfos []models.DailyDownloadInfo
	whichDayStr := ""
	if len(whichDay) > 0 {
		whichDayStr = whichDay[0]
	} else {
		nowTime := time.Now()
		whichDayStr = nowTime.Format("2006-01-02")
	}
	c.db.Where("supplier_name = ? AND public_ip = ? AND  which_day = ", supplierName, publicIP, whichDayStr).Find(&dailyDownloadInfos)

	if len(dailyDownloadInfos) == 0 {
		// 不存在
		return 0, nil
	}

	return dailyDownloadInfos[0].Count, nil
}

// AddDailyDownloadCount 根据字幕提供者的名称，今日下载计数的次数+1，仅仅统计次数，并不确认是哪个视频的字幕下载
func (c *CacheCenter) AddDailyDownloadCount(supplierName string, publicIP string, whichDay ...string) (int, error) {
	defer c.locker.Unlock()
	c.locker.Lock()

	dailyDownloadCount, err := c.GetDailyDownloadCount(supplierName, publicIP, whichDay...)
	if err != nil {
		return 0, err
	}

	whichDayStr := ""
	if len(whichDay) > 0 {
		whichDayStr = whichDay[0]
	} else {
		nowTime := time.Now()
		whichDayStr = nowTime.Format("2006-01-02")
	}

	dailyDownloadCount += 1
	dd := models.DailyDownloadInfo{
		SupplierName: supplierName,
		PublicIP:     publicIP,
		Count:        dailyDownloadCount,
		WhichDay:     whichDayStr,
	}
	c.db.Save(&dd)

	return dailyDownloadCount, nil
}
