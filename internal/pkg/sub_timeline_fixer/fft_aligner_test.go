package sub_timeline_fixer

import "testing"

func TestFFTAligner_fit(t *testing.T) {
	type args struct {
		refFloats []float64
		subFloats []float64
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "00", args: args{
			refFloats: []float64{1, 1, 1, 1, -1, -1, 1},
			subFloats: []float64{-1, 1, 1, -1, -1, 1, -1},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := FFTAligner{}
			f.fit(tt.args.refFloats, tt.args.subFloats)
		})
	}
}
