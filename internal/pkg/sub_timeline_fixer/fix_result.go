package sub_timeline_fixer

type FixResult struct {
	OldMean float64
	OldSD   float64
	NewMean float64
	NewSD   float64
	Per     float64 // 占比
}
