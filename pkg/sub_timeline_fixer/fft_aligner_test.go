package sub_timeline_fixer

import (
	"gonum.org/v1/gonum/floats/scalar"
	"testing"
)

func TestFFTAligner_Fit(t *testing.T) {
	type args struct {
		refFloats []float64
		subFloats []float64
	}
	tests := []struct {
		name      string
		args      args
		wantIndex int
		wantScore float64
	}{
		{name: "3-4", args: args{
			refFloats: []float64{1, 1, 1, 1, 1, -1, 1},
			subFloats: []float64{1, 1, -1, 1},
		}, wantIndex: 3, wantScore: 4},
		{name: "3-5", args: args{
			refFloats: []float64{0, 1, 1, 1, 1, -1, -1, 1},
			subFloats: []float64{1, 1, -1, -1, 1},
		}, wantIndex: 3, wantScore: 5},
	}
	const tol = 1e-10
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFFTAligner(2, 2)
			index, score := f.Fit(tt.args.refFloats, tt.args.subFloats)
			if index != tt.wantIndex {
				t.Errorf("Fit() wantIndex = %v, want %v", index, tt.wantIndex)
			}
			if scalar.EqualWithinAbsOrRel(score, tt.wantScore, tol, tol) == false {
				t.Errorf("Fit() wantScore = %v, want %v", score, tt.wantScore)
			}
		})
	}
}
