package shooter

import (
	"testing"
)

func TestNewSupplier(t *testing.T) {
	//movie1 := "X:\\电影\\消失爱人 (2016)\\消失爱人 (2016) 720p AAC.rmvb"
	//movie2 := "X:\\电影\\机动战士Z高达：星之继承者 (2005)\\机动战士Z高达：星之继承者 (2005) 1080p TrueHD.mkv"
	standard1 := "X:\\连续剧\\The Bad Batch\\Season 1\\The Bad Batch - S01E01 - Aftermath WEBDL-1080p.mkv"
	shooter := NewSupplier()
	outList, err := shooter.GetSubListFromFile(standard1, "")
	if err != nil {
		t.Error(err)
	}
	println(outList)

	for i, sublist := range outList {
		println(i, sublist.Language, sublist.Rate, sublist.Vote, sublist.FileUrl)
	}
}
