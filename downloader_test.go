package main

import (
	"github.com/allanpk716/ChineseSubFinder/common"
	"testing"
)

func TestDownloader_DownloadSub4Movie(t *testing.T) {
	var err error
	//dirRoot := "X:\\电影\\Spiral From the Book of Saw (2021)"
	//dirRoot := "X:\\电影\\Oslo (2021)"
	//dirRoot := "X:\\电影\\The Devil All the Time (2020)"
	//dirRoot := "X:\\电影\\21座桥 (2019)"
	dirRoot := "X:\\电影\\送你一朵小红花 (2020)"
	//dirRoot := "X:\\电影\\冰海陷落 (2018)"

	dl := NewDownloader(common.ReqParam{
		SaveMultiSub: true,
	})
	err = dl.DownloadSub4Movie(dirRoot)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDownloader_DownloadSub4Series(t *testing.T) {
	var err error
	//dirRoot := "X:\\连续剧\\隐秘的角落 (2020)"
	//dirRoot := "X:\\连续剧\\The Bad Batch"
	dirRoot := "X:\\连续剧\\Why Women Kill"

	dl := NewDownloader(common.ReqParam{
		SaveMultiSub: true,
	})
	err = dl.DownloadSub4Series(dirRoot)
	if err != nil {
		t.Fatal(err)
	}
}