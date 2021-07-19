package hot_fix

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"path"
	"path/filepath"
	"testing"
)

func TestHotFix001_GetKey(t *testing.T) {
	hf001 := NewHotFix001("", "")
	if hf001.GetKey() != "001" {
		t.Fatal("GetKey() != 001")
	}
}

func TestHotFix001_Process(t *testing.T) {
	testDataPath := "..\\..\\..\\TestData\\hotfix\\001"
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
	hf001 := NewHotFix001(testMovieDir, testSeriesDir)
	newSubFileName, errFiles, err := hf001.Process()
	if err != nil {
		for _, file := range errFiles {
			println("rename error:", file)
		}
		t.Fatal("Process ", err.Error())
	}
	if len(newSubFileName) < 1 {
		t.Fatal("hf001.Process() not file processed")
	}

	// 检查修复的结果是否符合预期
	var newSubFileNameMap = make(map[string]int)
	for i, s := range newSubFileName {
		if pkg.IsFile(s) == false {
			t.Fatal("renamed file not found:", s)
		}
		newSubFileNameMap[filepath.Base(s)] = i
	}

	// 21座桥 (2019) 720p AAC.chs[subhd].ass
	// 21座桥 (2019) 720p AAC.chs_en[zimuku].ass
	// 无罪之最 - S01E01 - 重建生活.chs[shooter].ass
	// 无罪之最 - S01E01 - 重建生活.chs[subhd].ass
	// 无罪之最 - S01E01 - 重建生活.chs[zimuku].ass
	// Loki - S01E01 - Glorious Purpose WEBDL-1080p Proper.chs_en[zimuku].srt
	// Loki - S01E01 - Glorious Purpose WEBDL-1080p Proper.cht_en[shooter].ass
	var checkResults = []string {
		"21座桥 (2019) 720p AAC.chinese(简,subhd).ass",
		"21座桥 (2019) 720p AAC.chinese(简英,zimuku).ass",
		"无罪之最 - S01E01 - 重建生活.chinese(简,shooter).ass",
		"无罪之最 - S01E01 - 重建生活.chinese(简,subhd).ass",
		"无罪之最 - S01E01 - 重建生活.chinese(简,zimuku).ass",
		"Loki - S01E01 - Glorious Purpose WEBDL-1080p Proper.chinese(简英,zimuku).srt",
		"Loki - S01E01 - Glorious Purpose WEBDL-1080p Proper.chinese(繁英,shooter).ass",
	}

	if len(newSubFileName) != len(checkResults) {
		t.Fatal("newSubFileName.len != checkResults.len")
	}

	for _, result := range checkResults {
		_, bok := newSubFileNameMap[result]
		if bok == false {
			t.Fatal("renamed file name not fit:", result)
		}
	}
}
