package model

import "testing"

func TestGetVideoInfoFromIMDB(t *testing.T) {
	//imdbID := "tt12708542" // 星球大战：残次品
	//imdbID := "tt7016936"	// 杀死伊芙
	//imdbID := "tt2990738" 	// 恐怖直播
	//imdbID := "tt3032476" 	// 风骚律师
	imdbID := "tt6468322" 	// 纸钞屋
	imdbInfo, err := GetVideoInfoFromIMDB(imdbID)
	if err != nil {
		t.Fatal(err)
	}
	println(imdbInfo.Name, imdbInfo.Year, imdbInfo.ID)
}