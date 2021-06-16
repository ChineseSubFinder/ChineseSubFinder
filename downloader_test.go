package main

import (
	"github.com/allanpk716/ChineseSubFinder/common"
	"testing"
)

func TestDownloader_DownloadSub(t *testing.T) {
	var err error
	//dirRoot := "X:\\电影\\Spiral From the Book of Saw (2021)"
	//dirRoot := "X:\\电影\\Oslo (2021)"
	//dirRoot := "X:\\电影\\The Devil All the Time (2020)"
	dirRoot := "X:\\电影\\冰海陷落 (2018)"

	dl := NewDownloader(common.ReqParam{
		SaveMultiSub: true,
	})
	err = dl.DownloadSub4Movie(dirRoot)
	if err != nil {
		t.Fatal(err)
	}
}