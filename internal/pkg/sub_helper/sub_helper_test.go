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

func TestGetVADInfoFeatureFromSub(t *testing.T) {

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

	baseUnitList, err := GetVADInfoFeatureFromSub(infoBase, FrontAndEndPerBase, 10000, true)
	if err != nil {
		t.Fatal(err)
	}
	baseUnit := baseUnitList[0]
	// Src，截取的部分要小于 Base 的部分
	srcUnitList, err := GetVADInfoFeatureFromSub(infoSrc, FrontAndEndPerBase, 10000, true)
	if err != nil {
		t.Fatal(err)
	}
	srcUnit := srcUnitList[0]

	if len(baseUnit.VADList) != len(srcUnit.VADList) {

		t.Fatal(fmt.Sprintf("VAD List Base And Src Not Same Length, baseUnit.VADList Len = %v, srcUnit.VADList Len = %v",
			len(baseUnit.VADList), len(srcUnit.VADList)))
	}

	//for i := 0; i < len(baseUnit.VADList); i++ {
	//
	//}

	println("Done")
}

const FrontAndEndPerBase = 0

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

	baseSubUnits, err := GetVADInfosFromSub(infoBase, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	srcSubUnits, err := GetVADInfosFromSub(infoSrc, 0, 1)
	if err != nil {
		t.Fatal(err)
	}

	println(len(baseSubUnits))
	println(len(srcSubUnits))
}
