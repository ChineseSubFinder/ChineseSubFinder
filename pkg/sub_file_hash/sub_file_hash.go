package sub_file_hash

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"math"
	"os"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types/common"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/decode"
)

// Calculate 视频文件的唯一ID，支持内程序内的 Fake 蓝光视频文件路径
func Calculate(filePath string) (string, error) {

	bok, _, STREAMDir := decode.IsFakeBDMVWorked(filePath)
	if bok == true {
		bdmvBigFileFPath, err := getBDMVBigFileFPath(STREAMDir)
		if err != nil {
			return "", err
		}
		filePath = bdmvBigFileFPath
	}

	fp, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = fp.Close()
	}()

	stat, err := fp.Stat()
	if err != nil {
		return "", err
	}
	size := float64(stat.Size())

	if size < 0xF000 {
		return "", common.VideoFileIsTooSmall
	}

	samplePositions := [samplingPoints]int64{
		4 * 1024,
		int64(math.Floor(size / 4)),
		int64(math.Floor(size / 4 * 2)),
		int64(math.Floor(size / 4 * 3)),
		int64(size - 8*1024)}

	fullBlock := make([]byte, samplingPoints*onePointLen)
	index := 0

	for _, position := range samplePositions {

		//f, err := os.Create(filepath.Join("c:\\Tmp", fmt.Sprintf("%d", position)+".videoPart"))
		//if err != nil {
		//	return "", err
		//}

		oneBlock := make([]byte, onePointLen)
		_, err = fp.ReadAt(oneBlock, position)
		if err != nil {
			//_ = f.Close()
			return "", err
		}

		for _, b := range oneBlock {
			fullBlock[index] = b
			index++
		}

		//_, err = f.Write(oneBlock)
		//if err != nil {
		//	return "", err
		//}
		//_ = f.Close()
	}

	sum := sha256.Sum256(fullBlock)

	return fmt.Sprintf("%x", sum), nil
}

func getBDMVBigFileFPath(STREAMDir string) (string, error) {

	// 因为上面可以检测出是否是蓝光电影，这里就是找到蓝光中最大的一个流文件，然后交给后续进行特征提取
	pathSep := string(os.PathSeparator)
	maxFileFPath := ""
	var maxFileSize int64
	maxFileSize = 0
	files, err := os.ReadDir(STREAMDir)
	if err != nil {
		return "", err
	}
	for _, curFile := range files {

		fullPath := STREAMDir + pathSep + curFile.Name()

		if curFile.IsDir() {
			// 只关心这一个文件夹
			continue
		}
		// 文件
		info, err := curFile.Info()
		if err != nil {
			continue
		}
		if info.Size() > maxFileSize {
			maxFileSize = info.Size()
			maxFileFPath = fullPath
		}
	}

	if maxFileFPath == "" {
		return "", errors.New("getBDMVBigFileFPath no file found")
	}

	return maxFileFPath, nil
}

const (
	samplingPoints = 5
	onePointLen    = 4 * 1024
)

const checkHash = "f5176762cb6d62471d71511fdddeb7b80c9a7a8939fce5cf172100b4a404a048"
