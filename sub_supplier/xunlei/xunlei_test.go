package xunlei

import (
	"testing"
)

func TestGetList(t *testing.T) {

	xunlie := NewSupplier()
	outList, err := xunlie.GetSubListFromFile("X:\\电影\\Spiral From the Book of Saw (2021)\\Spiral From the Book of Saw (2021) WEBDL-1080p.mkv", "")
	if err != nil {
		t.Error(err)
	}
	println(outList)

	for i, sublist := range outList {
		println(i, sublist.Language, sublist.Rate, sublist.Vote, sublist.FileUrl)
	}
}
