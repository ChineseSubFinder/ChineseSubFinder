package settings

type LocalChromeSettings struct {
	Enabled             bool   `json:"enabled"`
	LocalChromeExeFPath string `json:"local_chrome_exe_f_path"`
}
