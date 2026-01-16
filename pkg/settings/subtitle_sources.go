package settings

type SubtitleSources struct {
	AssrtSettings AssrtSettings `json:"assrt_settings"`
}

func NewSubtitleSources() *SubtitleSources {
	return &SubtitleSources{}
}
