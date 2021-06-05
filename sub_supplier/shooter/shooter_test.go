package shooter

import (
	"testing"
)

func TestNewSupplier(t *testing.T) {

	shooter := NewSupplier()
	outList, err := shooter.GetSubListFromFile("X:\\电影\\消失爱人 (2016)\\消失爱人 (2016) 720p AAC.rmvb", "")
	if err != nil {
		t.Error(err)
	}
	println(outList)

	for i, sublist := range outList {
		println(i, sublist.Language, sublist.Rate, sublist.Vote, sublist.FileUrl)
	}
}
