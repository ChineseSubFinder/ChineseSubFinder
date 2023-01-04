package hot_fix

import (
	"path/filepath"
	"testing"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/log_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/unit_test_helper"
)

func TestHotFix001_GetKey(t *testing.T) {
	hf001 := NewHotFix001(log_helper.GetLogger4Tester(), []string{""}, []string{""})
	if hf001.GetKey() != "001" {
		t.Fatal("GetKey() != 001")
	}
}

func TestHotFix001_Process(t *testing.T) {
	testDataPath := unit_test_helper.GetTestDataResourceRootPath([]string{"hotfix", "001"}, 4, true)

	movieDir := "movies"
	seriesDir := "series"
	// 测试文件夹
	testMovieDir := filepath.Join(testDataPath, movieDir)
	testSeriesDir := filepath.Join(testDataPath, seriesDir)
	// 开始修复
	hf001 := NewHotFix001(log_helper.GetLogger4Tester(), []string{testMovieDir}, []string{testSeriesDir})
	outData, err := hf001.Process()
	outStruct := outData.(OutStruct001)
	if err != nil {
		for _, file := range outStruct.ErrFiles {
			println("rename error:", file)
		}
		t.Fatal("Process ", err.Error())
	}

	if len(outStruct.RenamedFiles) < 1 {
		t.Fatal("hf001.Process() not file processed")
	}

	// 检查修复的结果是否符合预期
	var newSubFileNameMap = make(map[string]int)
	for i, s := range outStruct.RenamedFiles {
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
	var checkResults = []string{
		"21座桥 (2019) 720p AAC.chinese(简,subhd).ass",
		"21座桥 (2019) 720p AAC.chinese(简英,zimuku).ass",
		"无罪之最 - S01E01 - 重建生活.chinese(简,shooter).ass",
		"无罪之最 - S01E01 - 重建生活.chinese(简,subhd).ass",
		"无罪之最 - S01E01 - 重建生活.chinese(简,zimuku).ass",
		"Loki - S01E01 - Glorious Purpose WEBDL-1080p Proper.chinese(简英,zimuku).srt",
		"Loki - S01E01 - Glorious Purpose WEBDL-1080p Proper.chinese(繁英,shooter).ass",
	}

	if len(outStruct.RenamedFiles) != len(checkResults) {
		t.Logf("\n\nnewSubFileName.len %d != checkResults.len %d", len(outStruct.RenamedFiles), len(checkResults))
		t.FailNow()
	}

	for _, result := range checkResults {
		_, bok := newSubFileNameMap[result]
		if bok == false {
			t.Fatal("renamed file name not fit:", result)
		}
	}
}
