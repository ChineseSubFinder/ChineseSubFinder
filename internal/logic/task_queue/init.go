package task_queue

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/log_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_folder"
	"github.com/dgraph-io/badger/v3"
	"path/filepath"
)

func GetDb() *badger.DB {

	if dbBase == nil {
		var err error
		// 这边数据库会自动创建这个目录文件
		dbBase, err = badger.Open(badger.DefaultOptions(getDbName()))
		if err != nil {
			log_helper.GetLogger().Panicln("task_queue.GetDb()", err)
		}
	}
	return dbBase
}

func getDbName() string {
	return filepath.Join(my_folder.GetConfigRootDirFPath(), dbFileName)
}

const (
	dbFileName = "task_queue"
)

var dbBase *badger.DB
