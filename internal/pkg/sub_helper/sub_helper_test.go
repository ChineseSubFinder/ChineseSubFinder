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


	// ERROR: DeleteOneSeasonSubCacheFolder will create 
	// CSF-DebugThings/sub_helper
	// TODO: fix GetRootDebugFolder 

	testRootDir := "../../../TestData/sub_helper"
	err := DeleteOneSeasonSubCacheFolder(testRootDir)
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

	baseSubFile := filepath.FromSlash("../../../TestData/sub_helper/R&M-S05E10/Rick and Morty - S05E10 - Rickmurai Jack WEBRip-1080p.chinese(简,zimuku).default.srt")
	srcSubFile := filepath.FromSlash("../../../TestData/sub_helper/R&M-S05E10/Rick and Morty - S05E10 - Rickmurai Jack WEBRip-1080p.chinese(简英,zimuku).ass")

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

	if len(infoBase.DialoguesFilterEx) != len(infoSrc.DialoguesFilterEx) {
		t.Fatal(fmt.Sprintf("info Base And Src Parse Error, infoBase.DialoguesFilterEx Len = %v, infoSrc.DialoguesFilterEx Len = %v",
			len(infoBase.DialoguesFilterEx), len(infoSrc.DialoguesFilterEx)))
	}

	baseSubUnit, err := GetVADInfoFeatureFromSubNew(infoBase, FrontAndEndPerBase)
	if err != nil {
		t.Fatal(err)
	}
	srcSubUnit, err := GetVADInfoFeatureFromSubNew(infoSrc, FrontAndEndPerBase)
	if err != nil {
		t.Fatal(err)
	}
	if len(baseSubUnit.VADList) != len(srcSubUnit.VADList) {
		t.Fatal(fmt.Sprintf("info Base And Src Parse Error, infoBase.VADList Len = %v, infoSrc.VADList Len = %v",
			len(baseSubUnit.VADList), len(srcSubUnit.VADList)))
	}

	for i := 0; i < len(baseSubUnit.VADList); i++ {
		if baseSubUnit.VADList[i] != srcSubUnit.VADList[i] {
			println(fmt.Sprintf("base src VADList i=%v, not the same", i))
		}
	}
}

const FrontAndEndPerBase = 0
