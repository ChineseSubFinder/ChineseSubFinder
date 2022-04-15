package sub_file_hash

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/types/common"
	"math"
	"os"
)

// Calculate 视频文件的唯一ID
func Calculate(filePath string) (string, error) {

	h := sha1.New()
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

	h.Write(fullBlock)
	hashBytes := h.Sum(nil)

	return fmt.Sprintf("%x", md5.Sum(hashBytes)), nil
}

const (
	samplingPoints = 5
	onePointLen    = 4 * 1024
)

const checkHash = "f08d48a3e2cd6a02f9fd8ac92743dd3e"
