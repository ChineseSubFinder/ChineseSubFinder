package sort_things

import (
	"os"
	"sort"
	"time"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/decode"
)

type PathSlice struct {
	Path string
}
type PathSlices []PathSlice

func (a PathSlices) Len() int           { return len(a) }
func (a PathSlices) Less(i, j int) bool { return len(a[i].Path) < len(a[j].Path) }
func (a PathSlices) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// SortStringSliceByLength 排序得到匹配上的路径，最长的那个
func SortStringSliceByLength(m []string) PathSlices {
	p := make(PathSlices, len(m))
	i := 0
	for _, v := range m {
		p[i] = PathSlice{v}
		i++
	}
	sort.Sort(sort.Reverse(p))
	return p
}

// -----------------------------------------------------------------------------

// SortByModTime 根据文件的 Mod Time 进行排序，递减
func SortByModTime(fileList []string) []string {

	byModTime := make(ByModTime, 0)
	byModTime = append(byModTime, fileList...)
	sort.Sort(sort.Reverse(byModTime))

	return byModTime
}

type ByModTime []string

func (fis ByModTime) Len() int {
	return len(fis)
}

func (fis ByModTime) Swap(i, j int) {
	fis[i], fis[j] = fis[j], fis[i]
}

func (fis ByModTime) Less(i, j int) bool {

	aModTime := GetFileModTime(fis[i])
	bModTime := GetFileModTime(fis[j])

	return aModTime.Before(bModTime)
}

func GetFileModTime(fileFPath string) time.Time {

	if IsFile(fileFPath) == true {
		// 存在
		fi, err := os.Stat(fileFPath)
		if err != nil {
			return time.Time{}
		}

		return fi.ModTime()
	} else {
		// 不存在才需要考虑蓝光情况
		bok, idBDMVFPath, _ := decode.IsFakeBDMVWorked(fileFPath)
		if bok == false {
			// 也不是蓝光
			return time.Time{}
		}
		// 获取这个蓝光 ID BDMV 文件的时间
		fInfo, err := os.Stat(idBDMVFPath)
		if err != nil {
			return time.Time{}
		}
		return fInfo.ModTime()
	}
}

// -----------------------------------------------------------------------------

// IsFile 存在且是文件
func IsFile(filePath string) bool {
	s, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	return !s.IsDir()
}
