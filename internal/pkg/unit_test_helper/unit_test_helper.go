package unit_test_helper

import (
	"path/filepath"
)

// GetTestDataResourceRootPath 向上返回几层就能够到 ChineseSubFinder-TestData 同级目录，然后进入其中的 resourceFolderName 资源文件夹中
func GetTestDataResourceRootPath(resourceFolderNames []string, goBackTimes int) string {

	times := ""
	for i := 0; i < goBackTimes; i++ {
		times += oneBackTime
	}
	outPath := times + testResourceProjectName

	for _, name := range resourceFolderNames {
		outPath += "/" + name
	}

	outPath = filepath.FromSlash(outPath)

	return outPath
}

const oneBackTime = "../"
const testResourceProjectName = "ChineseSubFinder-TestData"
