package ass

import "sort"

type StyleNameInfo struct {
	Name  string
	Count int
}
type StyleNameInfos []StyleNameInfo

func (a StyleNameInfos) Len() int           { return len(a) }
func (a StyleNameInfos) Less(i, j int) bool { return a[i].Count < a[j].Count }
func (a StyleNameInfos) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func sortMapByValue(m map[string]int) StyleNameInfos {
	p := make(StyleNameInfos, len(m))
	i := 0
	for k, v := range m {
		p[i] = StyleNameInfo{k, v}
		i++
	}
	sort.Sort(sort.Reverse(p))
	return p
}
