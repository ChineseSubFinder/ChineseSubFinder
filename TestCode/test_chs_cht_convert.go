package TestCode

import "github.com/go-creed/sat"

func convertChsCht() {
	/*
		dicter.Read 转简体
		dicter.ReadReverse 转繁体
	*/
	dicter := sat.DefaultDict()
	sstr := "什麼"
	println(dicter.Read(sstr))
	println(dicter.ReadReverse(sstr))
	sstr = "什么"
	println(dicter.Read(sstr))
	println(dicter.ReadReverse(sstr))
	sstr = "簡繁轉換"
	println(dicter.Read(sstr))
	println(dicter.ReadReverse(sstr))
}
