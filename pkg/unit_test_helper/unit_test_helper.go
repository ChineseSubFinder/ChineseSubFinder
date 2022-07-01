package unit_test_helper

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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
		if IsDir(orgDir) == false {
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

		onePart := func() error {
			partF, err := os.Open(filepath.Join(videoPartsRootPath, name+ext))
			if err != nil {
				return err
			}
			defer func() {
				_ = partF.Close()
			}()
			partAll, err := io.ReadAll(partF)
			if err != nil {
				return err
			}
			int64Numb, err := strconv.ParseInt(name, 10, 64)
			if err != nil {
				return err
			}
			_, err = f.WriteAt(partAll, int64Numb)
			if err != nil {
				return err
			}
			return nil
		}

		err = onePart()
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

		onePart := func() error {
			partF, err := os.Open(filepath.Join(videoPartsRootPath, name+ext))
			if err != nil {
				return err
			}
			defer func() {
				_ = partF.Close()
			}()
			partAll, err := io.ReadAll(partF)
			if err != nil {
				return err
			}
			int64Numb, err := strconv.ParseInt(name, 10, 64)
			if err != nil {
				return err
			}
			_, err = f.WriteAt(partAll, int64Numb)
			if err != nil {
				return err
			}

			return nil
		}

		err = onePart()
		if err != nil {
			return "", err
		}
	}

	return outVideoFPath, nil
}

// GenerateCSFVideoFile 这里为 CSF 的接口专门生成一个视频文件，瑞克和莫蒂 (2013)\Season 5\S05E09 .mkv
func GenerateCSFVideoFile(videoPartsRootPath string) (string, error) {

	const videoSize int64 = 640302895
	const videoName = "S05E09.mkv"
	const ext = ".videoPart"
	partNames := []string{"4096", "160075723", "320151447", "480227171", "640294703"}

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
		一共有 5 个检测点
	*/
	for _, name := range partNames {

		onePart := func() error {
			partF, err := os.Open(filepath.Join(videoPartsRootPath, name+ext))
			if err != nil {
				return err
			}
			defer func() {
				_ = partF.Close()
			}()
			partAll, err := io.ReadAll(partF)
			if err != nil {
				return err
			}
			int64Numb, err := strconv.ParseInt(name, 10, 64)
			if err != nil {
				return err
			}
			_, err = f.WriteAt(partAll, int64Numb)
			if err != nil {
				return err
			}

			return nil
		}

		err = onePart()
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
	addString, _, _, _ := GetNowTimeString()
	// 测试数据的文件夹
	orgDir := filepath.Join(srcDir, "org")
	testDir := filepath.Join(srcDir, "test")

	if IsDir(testDir) == true {
		err := ClearFolderEx(testDir, 10)
		if err != nil {
			return "", err
		}
	}

	// 多加一层，这样在批量测试的时候才不会出错
	testDirEx := filepath.Join(testDir, addString)
	err := CopyDir(orgDir, testDirEx)
	if err != nil {
		return "", err
	}
	return testDirEx, nil
}

// IsDir 存在且是文件夹
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// CopyFile copies a single file from src to dst
func CopyFile(src, dst string) error {
	var err error
	var srcFd *os.File
	var dstFd *os.File
	var srcInfo os.FileInfo

	if srcFd, err = os.Open(src); err != nil {
		return err
	}
	defer func() {
		_ = srcFd.Close()
	}()

	if dstFd, err = os.Create(dst); err != nil {
		return err
	}
	defer func() {
		_ = dstFd.Close()
	}()

	if _, err = io.Copy(dstFd, srcFd); err != nil {
		return err
	}
	if srcInfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcInfo.Mode())
}

// CopyDir copies a whole directory recursively
func CopyDir(src string, dst string) error {
	var err error
	var fds []os.DirEntry
	var srcInfo os.FileInfo

	if srcInfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	if fds, err = os.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := filepath.Join(src, fd.Name())
		dstfp := filepath.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = CopyDir(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = CopyFile(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

// GetNowTimeString 获取当前的时间，没有秒
func GetNowTimeString() (string, int, int, int) {
	nowTime := time.Now()
	addString := fmt.Sprintf("%d-%d-%d", nowTime.Hour(), nowTime.Minute(), nowTime.Nanosecond())
	return addString, nowTime.Hour(), nowTime.Minute(), nowTime.Nanosecond()
}

// ClearFolderEx 清空文件夹，文件夹名称有特殊之处，Hour-min-Nanosecond 的命名方式
// 如果调用的时候，已存在的文件夹的时间 min < 5 那么则清理
func ClearFolderEx(folderFullPath string, overtime int) error {

	_, hour, minute, _ := GetNowTimeString()
	pathSep := string(os.PathSeparator)
	files, err := os.ReadDir(folderFullPath)
	if err != nil {
		return err
	}
	for _, curFile := range files {
		fullPath := folderFullPath + pathSep + curFile.Name()
		if curFile.IsDir() {

			parts := strings.Split(curFile.Name(), "-")
			if len(parts) == 3 {
				// 基本是符合了，倒是还是需要额外的判断是否时间超过了
				tmpHourStr := parts[0]
				tmpMinuteStr := parts[1]
				tmpHour, err := strconv.Atoi(tmpHourStr)
				if err != nil {
					// 如果不符合命名格式，直接删除
					err = os.RemoveAll(fullPath)
					if err != nil {
						return err
					}
					continue
				}
				tmpMinute, err := strconv.Atoi(tmpMinuteStr)
				if err != nil {
					// 如果不符合命名格式，直接删除
					err = os.RemoveAll(fullPath)
					if err != nil {
						return err
					}
					continue
				}
				// 判断时间
				if tmpHour != hour {
					// 如果不符合命名格式，直接删除
					err = os.RemoveAll(fullPath)
					if err != nil {
						return err
					}
					continue
				}
				// 超过 5 min
				if minute-overtime > tmpMinute {
					// 如果不符合命名格式，直接删除
					err = os.RemoveAll(fullPath)
					if err != nil {
						return err
					}
					continue
				}
			} else {
				// 如果不符合命名格式，直接删除
				err = os.RemoveAll(fullPath)
				if err != nil {
					return err
				}
			}
		} else {
			// 这里就是文件了
			err = os.Remove(fullPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

const oneBackTime = "../"
const testResourceProjectName = "ChineseSubFinder-TestData"
