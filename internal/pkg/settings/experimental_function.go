package settings

// ExperimentalFunction 实验性功能
type ExperimentalFunction struct {
	AutoChangeSubEncode AutoChangeSubEncode `json:"auto_change_sub_encode"`
}

func NewExperimentalFunction() *ExperimentalFunction {
	return &ExperimentalFunction{}
}
