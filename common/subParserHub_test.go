package common

import (
	"github.com/allanpk716/ChineseSubFinder/sub_parser/ass"
	"github.com/allanpk716/ChineseSubFinder/sub_parser/srt"
	"testing"
)

func TestSubParserHub_DetermineFileTypeFromFile(t *testing.T) {

	//filePath := "C:\\tmp\\[zimuku]_0_Spiral.From.the.Book.of.Saw.2021.1080p.WEBRip.x264-RARBG.eng.srt"
	filePath := "C:\\tmp\\[zimuku]_0_Spiral.From.the.Book.of.Saw.2021.1080p.WEBRip.x264-RARBG.chi.srt"
	p := NewSubParserHub(ass.NewParser(), srt.NewParser())
	subFileInfo, err := p.DetermineFileTypeFromFile(filePath)
	if err != nil {
		t.Fatal(err)
	}
	println(subFileInfo.Name, subFileInfo.FromWhereSite, subFileInfo.Lang.String(), subFileInfo.Ext)
}
