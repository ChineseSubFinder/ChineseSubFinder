package sub_supplier

type SubInfo struct {
	Name 		string `json:"name"`
	Language 	string `json:"language"`
	Rate 		string `json:"rate"`
	FileUrl     string `json:"file-url"`
	Vote    	int64  `json:"vote"`
	Offset  	int64  `json:"offset"`
	Ext			string `json:"ext"`		// 字幕文件的后缀名带点，有可能是直接能用的字幕文件，也可能是压缩包
}

func NewSubInfo(name string, language string, rate string, fileUrl string, vote int64, offset int64, ext string) *SubInfo {
	return &SubInfo{Name: name, Language: language, Rate: rate, FileUrl: fileUrl, Vote: vote, Offset: offset, Ext: ext}
}