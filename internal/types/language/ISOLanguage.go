package language

// 中文语言编码变种
// 参考 https://en.wikipedia.org/wiki/Chinese_Wikipedia#Automatic_conversion_between_traditional_and_simplified_Chinese_characters
const (
	ChineseISO_Hans = "zh-hans" // 简体
	ChineseISO_Hant = "zh-hant" // 繁體
	ChineseISO_CN   = "zh-cn"   // 大陆简体
	ChineseISO_TW   = "zh-tw"   // 臺灣正體
	ChineseISO_SG   = "zh-sg"   // 新加坡简体/马新简体
	ChineseISO_MY   = "zh-my"   // 大马简体
	ChineseISO_HK   = "zh-hk"   // 香港繁體
	ChineseISO_MO   = "zh-mo"   // 澳門繁體
)

// 参考 https://en.wikipedia.org/wiki/List_of_ISO_639-1_codes 标准
const (
	ISO_639_1_Chinese  = "zh"
	ISO_639_1_English  = "en"
	ISO_639_1_Korean   = "ko"
	ISO_639_1_Japanese = "ja"
)

const (
	ISO_639_2T_Chinese  = "zho"
	ISO_639_2T_English  = "eng"
	ISO_639_2T_Korean   = "kor"
	ISO_639_2T_Japanese = "jpn"
)

const (
	ISO_639_2B_Chinese  = "chi"
	ISO_639_2B_English  = "eng"
	ISO_639_2B_Korean   = "kor"
	ISO_639_2B_Japanese = "jpn"
)

const (
	ISO_639_3_Chinese  = "zho"
	ISO_639_3_English  = "eng"
	ISO_639_3_Korean   = "kor"
	ISO_639_3_Japanese = "jpn"
)

const (
	RegISORuleFront = `(?mi)\.(`
	RegISORuleEnd   = `)(\.\S+)`
)
