package sub_timeline_fixer

import (
	"testing"
)

func Test_searchBackUpSubFile(t *testing.T) {

	dir := "X:\\连续剧"
	files, err := searchBackUpSubFile(dir)
	if err != nil {
		t.Fatal(err)
	}
	println(len(files))
}

func TestRestore(t *testing.T) {

	err := Restore("X:\\电影", "X:\\连续剧")
	if err != nil {
		t.Fatal(err)
	}
}
