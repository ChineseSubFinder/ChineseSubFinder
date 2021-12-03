package sub_helper

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/ass"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/srt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_parser_hub"
	"path/filepath"
	"testing"
)

func TestDeleteOneSeasonSubCacheFolder(t *testing.T) {

	testDataPath := "../../../TestData/sub_helper"
	testRootDir, err := my_util.CopyTestData(testDataPath)
	if err != nil {
		t.Fatal(err)
	}
	err = DeleteOneSeasonSubCacheFolder(testRootDir)
	if err != nil {
		t.Fatal(err)
	}
	if my_util.IsDir(filepath.Join(testRootDir, "Sub_S1E0")) == true {
		t.Fatal("Sub_S1E0 not delete")
	}
}

func TestGetVADInfosFromSub(t *testing.T) {

	// 这两个字幕是一样的，只不过是格式不同而已
	subParserHub := sub_parser_hub.NewSubParserHub(ass.NewParser(), srt.NewParser())
	baseSubFile := "C:\\Tmp\\Rick and Morty - S05E01\\英_2.srt"
	srcSubFile := "C:\\Tmp\\Rick and Morty - S05E01\\英_2.ass"

	bFind, infoBase, err := subParserHub.DetermineFileTypeFromFile(baseSubFile)
	if err != nil {
		t.Fatal(err)
	}
	if bFind == false {
		t.Fatal("sub not match")
	}
	bFind, infoSrc, err := subParserHub.DetermineFileTypeFromFile(srcSubFile)
	if err != nil {
		t.Fatal(err)
	}
	if bFind == false {
		t.Fatal("sub not match")
	}

	if len(infoBase.DialoguesEx) != len(infoSrc.DialoguesEx) {
		t.Fatal(fmt.Sprintf("info Base And Src Parse Error, infoBase.DialoguesEx Len = %v, infoSrc.DialoguesEx Len = %v",
			len(infoBase.DialoguesEx), len(infoSrc.DialoguesEx)))
	}

	baseSubUnits, err := GetVADInfosFromSub(infoBase, FrontAndEndPerBase, 1)
	if err != nil {
		t.Fatal(err)
	}
	baseSubUnit := baseSubUnits[0]
	srcSubUnits, err := GetVADInfosFromSub(infoSrc, FrontAndEndPerBase, 1)
	if err != nil {
		t.Fatal(err)
	}
	srcSubUnit := srcSubUnits[0]
	if len(baseSubUnit.VADList) != len(srcSubUnit.VADList) {
		t.Fatal(fmt.Sprintf("info Base And Src Parse Error, infoBase.VADList Len = %v, infoSrc.VADList Len = %v",
			len(baseSubUnit.VADList), len(srcSubUnit.VADList)))
	}

	for i := 0; i < len(baseSubUnit.VADList); i++ {
		if baseSubUnit.VADList[i] != srcSubUnit.VADList[i] {
			println(fmt.Sprintf("base src VADList i=%v, not the same", i))
		}
	}

	println(len(baseSubUnits))
	println(len(srcSubUnits))
}

const FrontAndEndPerBase = 0
