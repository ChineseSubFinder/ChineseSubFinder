package task_queue

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/xujiajun/nutsdb"
	"time"
)

// GetDailyDownloadCount 根据字幕提供者的名称，获取今日下载计数的次数
func GetDailyDownloadCount(supplierName string) (int, error) {

	nowTime := time.Now()
	// 今天剩余的时间（s）
	KeyName := supplierName + "_" + nowTime.Format("2006-01-02")

	outCount := 0
	err := GetDb().Update(
		func(tx *nutsdb.Tx) error {
			var err error
			outCount, err = getDailyDownloadCount(tx, KeyName)
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

func getDailyDownloadCount(tx *nutsdb.Tx, KeyName string) (int, error) {
	outCount := 0
	key := []byte(KeyName)
	// 先判断 key 是否存在
	ok, err := tx.SHasKey(BucketNamePrefixSupplierDailyDownloadCounter, key)
	if err != nil {
		return 0, err
	}
	if ok == true {
		// 存在
		// 不存在，说明今天的还没有使用，那么就需要新建赋值
		if e, err := tx.Get(BucketNamePrefixSupplierDailyDownloadCounter, key); err != nil {
			return 0, err
		} else {
			outCount, err = my_util.BytesToInt(e.Value)
			if err != nil {
				return 0, err
			}
		}
	} else {
		// 每日的计数器从 0 开始
		value, err := my_util.IntToBytes(0)
		if err != nil {
			return 0, err
		}
		restOfDaySecond := my_util.GetRestOfDaySec()
		// 不存在，说明今天的还没有使用，那么就需要新建赋值
		if err = tx.Put(BucketNamePrefixSupplierDailyDownloadCounter, key, value, uint32(restOfDaySecond.Seconds())); err != nil {
			return 0, err
		}

		return outCount, nil
	}

	return outCount, nil
}

// AddDailyDownloadCount 根据字幕提供者的名称，今日下载计数的次数+1，TTL 多加 5s 确保今天过去
func AddDailyDownloadCount(supplierName string) (int, error) {

	nowTime := time.Now()
	// 今天剩余的时间（s）
	restOfDaySecond := my_util.GetRestOfDaySec()
	KeyName := supplierName + "_" + nowTime.Format("2006-01-02")
	dailyDownloadCount := 0
	err := GetDb().Update(
		func(tx *nutsdb.Tx) error {

			var err error
			key := []byte(KeyName)
			dailyDownloadCount, err = getDailyDownloadCount(tx, KeyName)
			if err != nil {
				return err
			}
			value, err := my_util.IntToBytes(dailyDownloadCount + 1)
			if err != nil {
				return err
			}
			// 因为已经查询了一次，确保一定存在，所以直接更新+1，TTL 多加 5s 确保今天过去
			if err = tx.Put(BucketNamePrefixSupplierDailyDownloadCounter, key, value, uint32(restOfDaySecond.Seconds())+5); err != nil {
				return err
			}
			return nil
		})
	if err != nil {
		return 0, err
	}

	return dailyDownloadCount, nil
}
