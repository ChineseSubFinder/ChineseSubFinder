package hot_fix

import (
	"path/filepath"
	"testing"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types"

	"github.com/ChineseSubFinder/ChineseSubFinder/internal/dao"
	"github.com/ChineseSubFinder/ChineseSubFinder/internal/models"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/log_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/unit_test_helper"
)

func TestHotFixProcess(t *testing.T) {

	// 先删除 db
	err := dao.DeleteDbFile()
	if err != nil {
		t.Fatal(err)
	}
	testDataPath := unit_test_helper.GetTestDataResourceRootPath([]string{"hotfix", "001"}, 4, true)
	movieDir := "movies"
	seriesDir := "series"
	// 测试文件夹
	testMovieDir := filepath.Join(testDataPath, movieDir)
	testSeriesDir := filepath.Join(testDataPath, seriesDir)
	// 开始修复
	err = HotFixProcess(log_helper.GetLogger4Tester(), types.HotFixParam{
		MovieRootDirs:  []string{testMovieDir},
		SeriesRootDirs: []string{testSeriesDir},
	})
	if err != nil {
		t.Fatal(err)
	}
	// 判断数据库的标记是否正确
	hotfixResult := models.HotFix{}
	result := dao.GetDb().Where("key = ?", "001").First(&hotfixResult)
	if result.Error != nil {
		t.Fatal(result.Error)
	}
}
