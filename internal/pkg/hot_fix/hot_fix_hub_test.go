package hot_fix

import (
	"github.com/allanpk716/ChineseSubFinder/internal/dao"
	"github.com/allanpk716/ChineseSubFinder/internal/models"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"github.com/allanpk716/ChineseSubFinder/internal/types"
	"path"
	"testing"
)

func TestHotFixProcess(t *testing.T) {

	// 先删除 db
	err := dao.DeleteDbFile()
	if err != nil {
		t.Fatal(err)
	}
	testDataPath := "../../../TestData/hotfix/001"
	movieDir := "movies"
	seriesDir := "series"
	testRootDir, err := pkg.CopyTestData(testDataPath)
	if err != nil {
		t.Fatal(err)
	}
	// 测试文件夹
	testMovieDir := path.Join(testRootDir, movieDir)
	testSeriesDir := path.Join(testRootDir, seriesDir)
	// 开始修复
	err = HotFixProcess(types.HotFixParam{
		MovieRootDir:  testMovieDir,
		SeriesRootDir: testSeriesDir,
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
