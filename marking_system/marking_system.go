package marking_system

// MarkingSystem 评价系统，解决字幕排序优先级问题
type MarkingSystem struct {
	subSiteSequence []string	// 网站的优先级，从高到低
}

func NewMarkingSystem(subSiteSequence []string) *MarkingSystem {
	return &MarkingSystem{subSiteSequence: subSiteSequence}
}

func (m MarkingSystem) SelectOneSubFile() {

}