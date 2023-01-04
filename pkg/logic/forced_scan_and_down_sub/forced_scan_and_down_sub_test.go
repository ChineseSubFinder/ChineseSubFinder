package forced_scan_and_down_sub

import (
	"os"
	"testing"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"
)

func TestCheckSpeFile(t *testing.T) {

	if getSpeFileName() == "" {
		t.Fatal("this OS not support")
	}

	ff, err := os.Create(getSpeFileName())
	if err != nil {
		t.Fatal(err)
	}
	_, err = ff.WriteString("123")
	if err != nil {
		t.Fatal(err)
	}
	err = ff.Close()
	if err != nil {
		t.Fatal(err)
	}

	got, err := CheckSpeFile()
	if err != nil {
		t.Fatal(err)
	}

	if got == false || pkg.IsFile(getSpeFileName()) == true {
		t.Fatal("CheckSpeFile fatal")
	}
}
