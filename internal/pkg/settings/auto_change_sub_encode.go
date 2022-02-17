package settings

type AutoChangeSubEncode struct {
	Enable        bool `json:"enable"`
	DesEncodeType int  `json:"des_encode_type"` // 默认 0 是 UTF-8，1 是 GBK
}
