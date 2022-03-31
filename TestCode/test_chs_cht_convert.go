package TestCode

import "github.com/go-creed/sat"

func convertChsCht() {
	/*
		dicter.Read 转简体
		dicter.ReadReverse 转繁体
	*/
	dicter := sat.DefaultDict()
	println("---------------------")
	sstr := "什麼sdg as 123"
	println(sstr)
	// 转换到 简体
	println(dicter.Read(sstr))
	// 转换到 繁体
	println(dicter.ReadReverse(sstr))
	println("---------------------")
	sstr = "11什asd么56"
	println(sstr)
	println(dicter.Read(sstr))
	println(dicter.ReadReverse(sstr))
	println("---------------------")
	sstr = "asd簡繁as25轉s換66"
	println(sstr)
	println(dicter.Read(sstr))
	println(dicter.ReadReverse(sstr))
	println("---------------------")
}
