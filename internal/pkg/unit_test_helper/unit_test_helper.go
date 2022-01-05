package unit_test_helper

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"path/filepath"
)

// GetTestDataResourceRootPath 向上返回几层就能够到 ChineseSubFinder-TestData 同级目录，然后进入其中的 resourceFolderName 资源文件夹中
func GetTestDataResourceRootPath(resourceFolderNames []string, goBackTimes int, userCopyData bool) string {

	times := ""
	for i := 0; i < goBackTimes; i++ {
		times += oneBackTime
	}
	outPath := times + testResourceProjectName

	for _, name := range resourceFolderNames {
		outPath += "/" + name
	}

	outPath = filepath.FromSlash(outPath)

	if userCopyData == true {
		// 想要 copy org 中的数据到 test 中去处理
		orgDir := filepath.Join(outPath, "org")
		if my_util.IsDir(orgDir) == false {
			// 如果没有发现 org 文件夹，就返回之前的路径即可
			return outPath
		}
		// 如果发现有，那启动 copy 的操作
		testDataPath, err := copyTestData(outPath)
		if err != nil {
			return outPath
		}
		return filepath.FromSlash(testDataPath)
	}

	return outPath
}

// copyTestData 单元测试前把测试的数据 copy 一份出来操作，src 目录中默认应该有一个 org 原始数据文件夹，然后需要复制一份 test 文件夹出来
func copyTestData(srcDir string) (string, error) {
	// 测试数据的文件夹
	orgDir := filepath.Join(srcDir, "org")
	testDir := filepath.Join(srcDir, "test")

	if my_util.IsDir(testDir) == true {
		err := my_util.ClearFolder(testDir)
		if err != nil {
			return "", err
		}
	}

	err := my_util.CopyDir(orgDir, testDir)
	if err != nil {
		return "", err
	}
	return testDir, nil
}

const oneBackTime = "../"
const testResourceProjectName = "ChineseSubFinder-TestData"
