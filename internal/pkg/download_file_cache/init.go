package download_file_cache

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_folder"
	"github.com/dgraph-io/badger/v3"
	"path/filepath"
	"time"
)

func GetDb() *badger.DB {

	if dbBase == nil {
		var err error
		opt := badger.DefaultOptions(getDbName())
		// 1MB
		opt.ValueLogFileSize = 1 << 20
		// 10 MB
		opt.MemTableSize = 10 << 20
		// 这边数据库会自动创建这个目录文件
		dbBase, err = badger.Open(opt)
		if err != nil {
			panic(err)
		}

		go badgerGC(dbBase)
	}
	return dbBase
}

func DelDb() error {

	if dbBase != nil {
		_ = dbBase.Close()
	}
	return my_folder.ClearFolder(getDbName())
}

func getDbName() string {
	return filepath.Join(my_folder.GetConfigRootDirFPath(), dbFileName)
}

func badgerGC(_dbBase *badger.DB) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
	again:
		err := _dbBase.RunValueLogGC(0.5)
		if err == nil {
			goto again
		}
	}
}

const (
	dbFileName = "download_sub_cache"
)

var dbBase *badger.DB
