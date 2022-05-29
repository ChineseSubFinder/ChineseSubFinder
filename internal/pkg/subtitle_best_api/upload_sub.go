package subtitle_best_api

type UploadSubReq struct {
	SubSha256    string `json:"sub_sha256"`     // 文件的 SHA256
	Season       int    `json:"season"`         // 如果对应的是电影则可能是 0，没有
	Episode      int    `json:"episode"`        // 如果对应的是电影则可能是 0，没有
	IsDouble     bool   `json:"is_double"`      // 是否是双语，基础是中文
	LanguageISO  string `json:"language_iso"`   // 字幕的语言，目标语言，就算是双语，中英，也应该是中文。ISO_639-1_codes 标准，见 ISOLanguage.go 文件，这里无法区分简体繁体
	MyLanguage   string `json:"my_language"`    // 这个是本程序定义的语言类型，见 my_language.go 文件
	ExtraPreName string `json:"extra_pre_name"` // 字幕额外的命名信息，指 Emby 字幕命名格式(简英,subhd)，的 subhd
	ImdbId       string `json:"imdb_id"`        // IMDB ID
	TmdbId       string `json:"tmdb_id"`        // TMDB ID，这里是这个剧集的 TMDB ID 不是这一集的哈
	VideoFeature string `json:"video_feature"`  // VideoFeature ID
	Year         int    `json:"year"`           // 年份,比如 2019、2022
}

type UploadSubReply struct {
	Status  int    `json:"status"` // 0 失败，1 成功，2 超过了上传时间，需要再此申请上传（AskForUploadReply）
	Message string `json:"message"`
}
