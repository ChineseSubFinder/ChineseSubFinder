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

// GenerateXunleiVideoFile 这里为 xunlei 的接口专门生成一个视频文件，手机 (2003) 720p Cooker.rmvb
func GenerateXunleiVideoFile(videoPartsRootPath string) (string, error) {

	const videoSize int64 = 640302895
	const videoName = "手机 (2003) 720p Cooker.rmvb"
	const ext = ".videoPart"
	partNames := []string{"0", "311177499", "933512018"}

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
		一共有 3 个检测点
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

	// 因为会出现，批量测试的需求，那么如果每次都进行一次清理，那么就会导致之前创建的被清理掉，测试用例失败
	// 可以简单的按时间来判断，如果当前时间与以及存在文件夹名称相差在 5min，那么就清理掉
	addString, _, _, _ := my_util.GetNowTimeString()
	// 测试数据的文件夹
	orgDir := filepath.Join(srcDir, "org")
	testDir := filepath.Join(srcDir, "test")

	if my_util.IsDir(testDir) == true {
		err := my_util.ClearFolderEx(testDir, 2)
		if err != nil {
			return "", err
		}
	}

	// 多加一层，这样在批量测试的时候才不会出错
	testDirEx := filepath.Join(testDir, addString)
	err := my_util.CopyDir(orgDir, testDirEx)
	if err != nil {
		return "", err
	}
	return testDirEx, nil
}

const oneBackTime = "../"
const testResourceProjectName = "ChineseSubFinder-TestData"
