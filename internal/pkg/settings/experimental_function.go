package settings

// ExperimentalFunction 实验性功能
type ExperimentalFunction struct {
	AutoChangeSubEncode AutoChangeSubEncode `json:"auto_change_sub_encode"`
	ChsChtChanger       ChsChtChanger       `json:"chs_cht_changer"`
}

func NewExperimentalFunction() *ExperimentalFunction {
	return &ExperimentalFunction{}
}
