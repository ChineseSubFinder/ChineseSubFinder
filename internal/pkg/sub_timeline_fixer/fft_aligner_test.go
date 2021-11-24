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
		name  string
		args  args
		want  int
		want1 float64
	}{
		{name: "00", args: args{
			refFloats: []float64{1, 1, 1, 1, 0, 0, 1},
			subFloats: []float64{0, 1, 1, 0, 0, 1, 0},
		}, want: 1, want1: 5},
	}
	const tol = 1e-10
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := FFTAligner{}
			got, got1 := f.Fit(tt.args.refFloats, tt.args.subFloats)
			if got != tt.want {
				t.Errorf("Fit() got = %v, want %v", got, tt.want)
			}
			if scalar.EqualWithinAbsOrRel(got1, tt.want1, tol, tol) {
				t.Errorf("Fit() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
