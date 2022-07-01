package frechet

import "math"

/*
	https://github.com/artpar/frechet
*/

// Point is used to represent curves
type Point struct {
	X float64
	Y float64
}

func euclideanDistance(p1 Point, p2 Point) float64 {
	dx := p2.X - p1.X
	dy := p2.Y - p1.Y
	return math.Sqrt(dx*dx + dy*dy)
}

func min(x float64, y float64, z float64) float64 {
	if x < y {
		return math.Min(x, z)
	}

	return math.Min(y, z)
}

// Frechet is a dynamic programming implementation calculating the frechet distance between the two curves c1 and c2.
func Frechet(c1 []Point, c2 []Point) float64 {
	I := len(c1)
	J := len(c2)
	runningMaxI := 0.0
	for i := 0; i < I; i++ {
		currentMin := 1e+09
		for j := 0; j < J; j++ {
			currDist := euclideanDistance(c1[i], c2[j])
			currentMin = math.Min(currentMin, currDist)
		}

		runningMaxI = math.Max(runningMaxI, currentMin)
	}

	runningMaxJ := 0.0
	for j := 0; j < J; j++ {
		currentMin := 1e+09
		for i := 0; i < I; i++ {
			currDist := euclideanDistance(c1[i], c2[j])
			currentMin = math.Min(currentMin, currDist)
		}

		runningMaxJ = math.Max(runningMaxJ, currentMin)
	}

	return math.Max(runningMaxI, runningMaxJ)
}
