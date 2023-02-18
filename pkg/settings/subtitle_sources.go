package settings

type SubtitleSources struct {
	AssrtSettings        AssrtSettings        `json:"assrt_settings"`
	SubtitleBestSettings SubtitleBestSettings `json:"subtitle_best_settings"`
}

func NewSubtitleSources() *SubtitleSources {
	return &SubtitleSources{}
}
