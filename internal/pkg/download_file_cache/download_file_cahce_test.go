package download_file_cache

import (
	"bytes"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/allanpk716/ChineseSubFinder/internal/types/language"
	"github.com/allanpk716/ChineseSubFinder/internal/types/supplier"
	"testing"
	"time"
)

func TestDownloadFileCache_Get(t *testing.T) {

	dfc := NewDownloadFileCache(settings.GetSettings())
	found, _, err := dfc.Get("asd")
	if err != nil {
		t.Error(err)
	}

	if found == true {
		t.Error("Found sub")
	}
}

func TestDownloadFileCache_Add(t *testing.T) {

	dfc := NewDownloadFileCache(settings.GetSettings())

	inSubInfo := supplier.NewSubInfo("local", 1, "testfile", language.ChineseSimple, "2esd12c31c123asecqwec", 3, 0, ".srt", []byte{1, 2, 3, 4})
	err := dfc.Add(inSubInfo)
	if err != nil {
		t.Error(err)
	}

	fileUID := inSubInfo.GetUID()

	found, gotSubInfo, err := dfc.Get(fileUID)
	if err != nil {
		t.Error(err)
	}

	if found == false {
		t.Error("Not Found sub")
	}

	if gotSubInfo.GetUID() != fileUID {
		t.Error("Wrong UID")
	}

	if gotSubInfo.FileUrl != inSubInfo.FileUrl {
		t.Error("Wrong FileUrl")
	}

	if bytes.Equal(gotSubInfo.Data, inSubInfo.Data) == false {
		t.Error("Wrong Data")
	}

}

func TestDownloadFileCache_AddTTL(t *testing.T) {

	nowSettings := settings.GetSettings()
	nowSettings.AdvancedSettings.DownloadFileCache.TTL = 3
	nowSettings.AdvancedSettings.DownloadFileCache.Unit = "second"
	dfc := NewDownloadFileCache(nowSettings)

	inSubInfo := supplier.NewSubInfo("local", 1, "testfile", language.ChineseSimple, "2esd12c31c123asecqwec", 3, 0, ".srt", []byte{1, 2, 3, 4})
	err := dfc.Add(inSubInfo)
	if err != nil {
		t.Error(err)
	}

	fileUID := inSubInfo.GetUID()

	time.Sleep(time.Second * 4)

	found, _, err := dfc.Get(fileUID)
	if err != nil {
		t.Error(err)
	}

	if found == true {
		t.Error("Found sub")
	}
}
