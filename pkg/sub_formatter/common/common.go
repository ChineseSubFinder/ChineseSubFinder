package common

const FormatterNameString_Normal = "normal formatter"
const FormatterNameString_Emby = "emby formatter"
const NoMatchFormatter = "No Match formatter"

type FormatterName int

const (
	Emby   FormatterName = iota // Emby 格式 xxx.chinese.(简,shooter).ass
	Normal                      // 常规  xxx.zh.ass
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
