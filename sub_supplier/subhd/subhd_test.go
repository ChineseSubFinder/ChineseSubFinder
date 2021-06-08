package subhd

import (
	"github.com/allanpk716/ChineseSubFinder/common"
	"testing"
)

func TestSupplier_GetSubListFromFile(t *testing.T) {

	movie1 := "X:\\电影\\消失爱人 (2016)\\消失爱人 (2016) 720p AAC.rmvb"
	//movie1 := "X:\\电影\\机动战士Z高达：星之继承者 (2005)\\机动战士Z高达：星之继承者 (2005) 1080p TrueHD.mkv"
	//movie1 := "X:\\连续剧\\The Bad Batch\\Season 1\\The Bad Batch - S01E01 - Aftermath WEBDL-1080p.mkv"
	shooter := NewSupplier(common.ReqParam{Topic: 3})
	outList, err := shooter.GetSubListFromFile(movie1)
	if err != nil {
		t.Error(err)
	}
	println(outList)

	for i, sublist := range outList {
		println(i, sublist.Name, sublist.Ext, sublist.Language.String(), sublist.Vote, sublist.FileUrl, len(sublist.Data))
	}
}

func TestSupplier_Simulation(t *testing.T) {
	shooter := NewSupplier()
	shooter.Simulation()
}