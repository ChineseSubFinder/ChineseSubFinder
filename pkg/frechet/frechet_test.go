package frechet

import (
	"testing"
)

func TestFrechet(t *testing.T) {
	curve1 := []Point{Point{X: 0, Y: 0}, Point{X: 1, Y: 1}, Point{X: 2, Y: 2}}
	curve2 := []Point{Point{X: 0, Y: 1}, Point{X: 1, Y: 2}, Point{X: 2, Y: 3}}

	dist := Frechet(curve1, curve2)
	if dist != 1.0 {
		t.Fatalf("%v != 1.0", dist)
	}
}

func BenchmarkFrechet(b *testing.B) {
	curve1 := []Point{}
	curve2 := []Point{}

	for i := 0; i < 1000; i++ {
		curve1 = append(curve1, Point{X: float64(i), Y: float64(i)})
		curve2 = append(curve2, Point{X: float64(i), Y: float64(i + 1)})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dist := Frechet(curve1, curve2)
		if dist != 1.0 {
			b.Fatalf("%v != 1.0", dist)
		}
	}
}
