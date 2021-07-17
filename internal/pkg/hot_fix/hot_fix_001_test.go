package hot_fix

import "testing"

func TestHotFix001_GetKey(t *testing.T) {
	hf001 := NewHotFix001("", "")
	if hf001.GetKey() != "001" {
		t.Fatal("GetKey() != 001")
	}
}

func TestHotFix001_Process(t *testing.T) {
	movieDir := "X:\\电影\\21座桥 (2019)"
	seriesDir := "X:\\连续剧\\无罪之最 (2021)"

	hf001 := NewHotFix001(movieDir, seriesDir)
	err := hf001.Process()
	if err != nil {
		t.Fatal("Process ", err.Error())
	}
}
