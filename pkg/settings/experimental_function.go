package settings

// ExperimentalFunction 实验性功能
type ExperimentalFunction struct {
	AutoChangeSubEncode  AutoChangeSubEncode  `json:"auto_change_sub_encode"`
	ChsChtChanger        ChsChtChanger        `json:"chs_cht_changer"`
	RemoteChromeSettings RemoteChromeSettings `json:"remote_chrome_settings"`
	ApiKeySettings       ApiKeySettings       `json:"api_key_settings"`
	LocalChromeSettings  LocalChromeSettings  `json:"local_chrome_settings"`
	ShareSubSettings     ShareSubSettings     `json:"share_sub_settings"`
}

func NewExperimentalFunction() *ExperimentalFunction {
	return &ExperimentalFunction{}
}
