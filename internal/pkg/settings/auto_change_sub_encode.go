package settings

type AutoChangeSubEncode struct {
	Enable        bool `json:"enable"`
	DesEncodeType int  `json:"des_encode_type"` // 默认 0 是 UTF-8，1 是 GBK
}

func (a AutoChangeSubEncode) GetDesEncodeType() string {
	if a.DesEncodeType == 0 {
		return "UTF-8"
	} else if a.DesEncodeType == 1 {
		return "GBK2312"
	} else {
		return "no support type"
	}
}
