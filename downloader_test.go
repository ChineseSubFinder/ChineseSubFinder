package ChineseSubFinder

import (
	"github.com/allanpk716/ChineseSubFinder/common"
	"github.com/allanpk716/ChineseSubFinder/sub_supplier/shooter"
	"testing"
)

func TestDownloader_searchFile(t *testing.T) {

	//dirRoot := "X:\\动漫\\EVA"
	dirRoot := "X:\\电影\\Spiral From the Book of Saw (2021)"

	dl := NewDownloader()
	files, err := dl.searchMatchedVideoFile(dirRoot)
	if err != nil {
		t.Fatal(err)
	}
	sp := shooter.NewSupplier()
	for i, file := range files {
		println(i, file)
		_, err := sp.ComputeFileHash(file)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestDownloader_DownloadSub(t *testing.T) {
	var err error
	dirRoot := "X:\\电影\\Spiral From the Book of Saw (2021)"

	dl := NewDownloader(common.ReqParam{DebugMode: true})
	err = dl.DownloadSub(dirRoot)
	if err != nil {
		t.Fatal(err)
	}
}