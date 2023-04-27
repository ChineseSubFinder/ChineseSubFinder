package common

const FormatterNameString_Normal = "normal formatter"
const FormatterNameString_Emby = "emby formatter"
const FormatterNameString_SampleAsVideoName = "sample as video name formatter"
const NoMatchFormatter = "No Match formatter"

type FormatterName int

const (
	Emby            FormatterName = iota // Emby 格式 xxx.chinese.(简,shooter).ass
	Normal                               // 常规  xxx.zh.ass
	SameAsVideoName                      // 与视频文件名称相同
)

func (f FormatterName) String() string {
	switch f {
	case Normal:
		return FormatterNameString_Normal
	case Emby:
		return FormatterNameString_Emby
	case SameAsVideoName:
		return FormatterNameString_SampleAsVideoName
	default:
		return NoMatchFormatter
	}
}
