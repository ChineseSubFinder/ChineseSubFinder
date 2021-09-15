package common

const FormatterNameString_Normal = "normal formatter"
const FormatterNameString_Emby = "emby formatter"
const NoMatchFormatter = "No Match formatter"

type FormatterName int

const (
	Normal FormatterName = iota // 常规  xxx.zh.ass
	Emby                        // Emby 格式 xxx.chinese.(简,shooter).ass
)

func (f FormatterName) String() string {
	switch f {
	case Normal:
		return FormatterNameString_Normal
	case Emby:
		return FormatterNameString_Emby
	default:
		return NoMatchFormatter
	}
}
