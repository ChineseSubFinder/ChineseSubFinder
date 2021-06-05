package sub_supplier

type SubInfo struct {
	Name 		string `json:"name"`
	Language 	string `json:"language"`
	Rate 		string `json:"rate"`
	FileUrl     string `json:"file-url"`
	Vote    	int64  `json:"vote"`
	Offset  	int64  `json:"offset"`
}

func NewSubInfo(name string, language string, rate string, fileUrl string, vote int64, offset int64) *SubInfo {
	return &SubInfo{Name: name, Language: language, Rate: rate, FileUrl: fileUrl, Vote: vote, Offset: offset}
}