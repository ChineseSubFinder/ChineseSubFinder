package zimuku

import (
	"testing"
)

func TestSupplier_GetSubListFromKeyword(t *testing.T) {

	//imdbId1 := "tt3228774"
	videoName := "黑白魔女库伊拉"
	s := Supplier{}
	subList, err := s.GetSubListFromKeyword(videoName, "")
	if err != nil {
		t.Error(err)
	}

	for _, info := range subList {
		println(info.Name)
	}
}
