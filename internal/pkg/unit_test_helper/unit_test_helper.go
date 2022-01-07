package unit_test_helper

import (
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"io"
	"os"
	"path/filepath"
	"strconv"
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

// GenerateShooterVideoFile 这里为 shooter 的接口专门生成一个视频文件，瑞克和莫蒂 (2013)\Season 5\S05E09 .mkv
func GenerateShooterVideoFile(videoPartsRootPath string) (string, error) {

	const videoSize int64 = 640302895
	const videoName = "S05E09.mkv"
	const ext = ".videoPart"
	partNames := []string{"4096", "213434298", "426868596", "640294703"}

	outVideoFPath := filepath.Join(videoPartsRootPath, videoName)

	f, err := os.Create(outVideoFPath)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = f.Close()
	}()

	if err := f.Truncate(videoSize); err != nil {
		return "", err
	}

	/*
		一共有 4 个检测点
	*/
	for _, name := range partNames {

		partF, err := os.Open(filepath.Join(videoPartsRootPath, name+ext))
		if err != nil {
			return "", err
		}
		partAll, err := io.ReadAll(partF)
		if err != nil {
			return "", err
		}
		int64Numb, err := strconv.ParseInt(name, 10, 64)
		if err != nil {
			return "", err
		}
		_, err = f.WriteAt(partAll, int64Numb)
		if err != nil {
			return "", err
		}
	}

	return outVideoFPath, nil
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
