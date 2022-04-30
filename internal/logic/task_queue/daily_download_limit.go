package task_queue

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/badger_err_check"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/dgraph-io/badger/v3"
	"time"
)

// GetDailyDownloadCount 根据字幕提供者的名称，获取今日下载计数的次数，仅仅统计次数，并不确认是那个视频的字幕下载
// whichDay nowTime.Format("2006-01-02")
func GetDailyDownloadCount(supplierName string, publicIP string, whichDay ...string) (int, error) {

	KeyName := ""
	if len(whichDay) > 0 {
		KeyName = supplierName + "_" + whichDay[0] + "_" + publicIP
	} else {
		nowTime := time.Now()
		KeyName = supplierName + "_" + nowTime.Format("2006-01-02") + "_" + publicIP
	}

	outCount := 0
	err := GetDb().View(
		func(tx *badger.Txn) error {
			var err error

			key := []byte(MergeBucketAndKeyName(BucketNamePrefixSupplierDailyDownloadCounter, KeyName))
			e, err := tx.Get(key)
			if err != nil {

				if badger_err_check.IsErrOk(err) == true {
					return nil
				}

				return err
			}
			valCopy, err := e.ValueCopy(nil)
			if err != nil {
				return err
			}
			outCount, err = my_util.BytesToInt(valCopy)
			if err != nil {
				return err
			}
			return nil
		})
	if err != nil {
		return 0, err
	}

	return outCount, nil
}

// AddDailyDownloadCount 根据字幕提供者的名称，今日下载计数的次数+1，仅仅统计次数，并不确认是那个视频的字幕下载。TTL 多加 5s 确保今天过去（暂时去除 TTL）
func AddDailyDownloadCount(supplierName string, publicIP string, whichDay ...string) (int, error) {

	nowTime := time.Now()
	// 今天剩余的时间（s）
	//restOfDaySecond := my_util.GetRestOfDaySec()
	KeyName := supplierName + "_" + nowTime.Format("2006-01-02") + "_" + publicIP

	dailyDownloadCount, err := GetDailyDownloadCount(supplierName, publicIP, whichDay...)
	if err != nil {
		return 0, err
	}

	err = GetDb().Update(
		func(tx *badger.Txn) error {

			var err error
			key := []byte(MergeBucketAndKeyName(BucketNamePrefixSupplierDailyDownloadCounter, KeyName))
			dailyDownloadCount += 1
			value, err := my_util.IntToBytes(dailyDownloadCount)
			if err != nil {
				return err
			}
			// 因为已经查询了一次，确保一定存在，所以直接更新+1，TTL 多加 5s 确保今天过去，暂时去除 TTL uint32(restOfDaySecond.Seconds())+5
			// .WithTTL(2 * time.Second)
			e := badger.NewEntry(key, value)
			err = tx.SetEntry(e)
			if err != nil {
				return err
			}
			return nil
		})
	if err != nil {
		return 0, err
	}

	return dailyDownloadCount, nil
}

// DelDailyDownloadCount 不是定版。这里需要搜索前缀去删除，因为引入了 Public IP 的分类
func DelDailyDownloadCount(supplierName string, whichDay ...string) error {

	KeyName := ""
	if len(whichDay) > 0 {
		KeyName = supplierName + "_" + whichDay[0]
	} else {
		nowTime := time.Now()
		KeyName = supplierName + "_" + nowTime.Format("2006-01-02")
	}

	err := GetDb().Update(
		func(tx *badger.Txn) error {

			var err error
			key := []byte(MergeBucketAndKeyName(BucketNamePrefixSupplierDailyDownloadCounter, KeyName))
			// 因为已经查询了一次，确保一定存在，所以直接更新+1，TTL 多加 5s 确保今天过去，暂时去除 TTL uint32(restOfDaySecond.Seconds())+5
			if err = tx.Delete(key); err != nil {
				return err
			}
			return nil
		})
	if err != nil {
		return err
	}

	return nil
}

// GetDailyVideoSubDownloadCount 今日有那些视频进行了字幕的下载
func GetDailyVideoSubDownloadCount() {

}

func getDailyVideoSubDownloadCount(tx *badger.Txn, KeyName string) {

}
