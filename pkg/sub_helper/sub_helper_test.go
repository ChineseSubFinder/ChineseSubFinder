package sub_helper

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/log_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_parser/ass"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/logic/sub_parser/srt"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/sub_parser_hub"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/unit_test_helper"
)

func TestDeleteOneSeasonSubCacheFolder(t *testing.T) {
	const testSerName = "XXX"
	const needDelFolderName = "Sub_S1E0"
	testRootDir := unit_test_helper.GetTestDataResourceRootPath([]string{"sub_helper", "org", needDelFolderName}, 4, false)
	desSerFullPath, err := pkg.GetDebugFolderByName([]string{testSerName})
	if err != nil {
		t.Fatal(err)
	}
	desSeasonFullPath, err := pkg.GetDebugFolderByName([]string{testSerName, filepath.Base(testRootDir)})
	if err != nil {
		t.Fatal(err)
	}
	err = pkg.CopyDir(testRootDir, desSeasonFullPath)
	if err != nil {
		t.Fatal(err)
	}
	err = DeleteOneSeasonSubCacheFolder(desSerFullPath)
	if err != nil {
		t.Fatal(err)
	}
	if pkg.IsDir(desSeasonFullPath) == true {
		t.Fatal("Sub_S1E0 not delete")
	}
}

func TestGetVADInfosFromSub(t *testing.T) {

	log := log_helper.GetLogger4Tester()
	// 这两个字幕是一样的，只不过是格式不同而已
	subParserHub := sub_parser_hub.NewSubParserHub(log, ass.NewParser(log), srt.NewParser(log))

	testRootDir := unit_test_helper.GetTestDataResourceRootPath([]string{"sub_helper", "org", "R&M-S05E10"}, 4, false)

	baseSubFile := filepath.Join(testRootDir, "Rick and Morty - S05E10 - Rickmurai Jack WEBRip-1080p.chinese(简,zimuku).default.srt")
	srcSubFile := filepath.Join(testRootDir, "Rick and Morty - S05E10 - Rickmurai Jack WEBRip-1080p.chinese(简英,zimuku).ass")

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

	if len(infoBase.Dialogues) != len(infoSrc.Dialogues) {
		t.Fatal(fmt.Sprintf("info Base And Src Parse Error, infoBase.DialoguesFilterEx Len = %v, infoSrc.DialoguesFilterEx Len = %v",
			len(infoBase.Dialogues), len(infoSrc.Dialogues)))
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
