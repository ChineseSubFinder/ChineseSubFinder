package subtitle_best_api

type Subtitle struct {
	SubSha256         string `json:"sub_sha256"`              // 文件的 SHA256
	IsDouble          bool   `json:"is_double"`               // 是否是双语，基础是中文
	MyLanguage        string `json:"my_language"`             // 这个是本程序定义的语言类型，见 my_language.go 文件
	LowTrust          bool   `json:"low_trust"`               // 是否是低信任的
	ExtraPreName      string `json:"extra_pre_name,optional"` //字幕额外的命名信息，指 Emby 字幕命名格式(简英,subhd)，的 subhd
	MatchVideoFeature bool   `json:"match_video_feature"`     // 是否匹配到 VideoFeature
	Ext               string `json:"ext"`                     // 字幕的后缀名
}
