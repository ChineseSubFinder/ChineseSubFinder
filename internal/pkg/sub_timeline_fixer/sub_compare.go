package sub_timeline_fixer

type SubCompare struct {
	baseIndexDict      map[int]int
	srcIndexDict       map[int]int
	baseIndexList      []int
	srcIndexList       []int
	maxCompareDialogue int
	baseNowIndex       int
	srcNowIndex        int
}

func NewSubCompare(maxCompareDialogue int) *SubCompare {
	sc := SubCompare{
		baseIndexDict:      make(map[int]int, 0),
		srcIndexDict:       make(map[int]int, 0),
		baseIndexList:      make([]int, 0),
		srcIndexList:       make([]int, 0),
		maxCompareDialogue: maxCompareDialogue,
		baseNowIndex:       -1,
		srcNowIndex:        -1,
	}
	return &sc
}

// Add 添加元素进来比较，这里有个细节，如果理论上需要判断是 OffsetIndex 是 1-5 ，那么如果 1 add了，2 add 失败的时候，是应该清理后再 add 2
// 还有一种情况，从 1-5，添加到 4 的时候false了，那么应该回退到 2 进行 add，而不是从 4 开始
func (s *SubCompare) Add(baseNowIndex, srcNowIndex int) bool {
	// 如果是第一次 Add 的话，就直接把后续需要匹配的 OffsetIndex 字典的 Key 信息建立好
	// 再次调用本方法的时候就是 check 是否需要加的 key 存在于 字典 Key 中即可
	if len(s.baseIndexDict) == 0 {
		// 第一次
		for i := 0; i < s.maxCompareDialogue; i++ {
			s.baseIndexDict[baseNowIndex+i] = i
			s.srcIndexDict[srcNowIndex+i] = i
			s.baseIndexList = append(s.baseIndexList, baseNowIndex+i)
			s.srcIndexList = append(s.srcIndexList, srcNowIndex+i)
		}
		s.baseNowIndex = baseNowIndex
		s.srcNowIndex = srcNowIndex
	}
	// 可以理解为第二次开始才进入这个逻辑
	// 判断是否是预计的顺序 OffsetIndex
	_, okBase := s.baseIndexDict[baseNowIndex]
	_, okSrc := s.srcIndexDict[srcNowIndex]
	if okBase == false || okSrc == false {
		// 一定要存在，因为必须是可期待的 OffsetIndex
		return false
	}
	// 上面的判断仅仅是确定这个 index 是期望的范围内的，而不能保证顺序
	// 需要在这里没进行一次判断成功后，这里需要判断进入顺序判断的逻辑
	if s.baseIndexList[0] != baseNowIndex || s.srcIndexList[0] != srcNowIndex {
		return false
	}
	// 移除数组的首个元素
	s.baseIndexList = s.baseIndexList[1:]
	s.srcIndexList = s.srcIndexList[1:]

	return true
}

// Check 是否 Add 的元素已经足够满足 maxCompareDialogue 的数量要求了
// 这里有个细节，如果理论上需要判断是 OffsetIndex 是 1-5 ，如果 add 5 check 的时候 false，那么应该清理后，回退到 2 进行 add，而不是 6 开始
func (s *SubCompare) Check() bool {
	if len(s.baseIndexList) == 0 && len(s.srcIndexList) == 0 {
		return true
	} else {
		return false
	}
}

func (s *SubCompare) Clear() {
	s.baseIndexDict = make(map[int]int, 0)
	s.srcIndexDict = make(map[int]int, 0)
	s.baseIndexList = make([]int, 0)
	s.srcIndexList = make([]int, 0)
	s.baseNowIndex = -1
	s.srcNowIndex = -1
}

func (s *SubCompare) GetStartIndex() (int, int) {
	return s.baseNowIndex, s.srcNowIndex
}
