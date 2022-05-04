package sort_things

import "sort"

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
