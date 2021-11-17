package dtw

import (
	"container/list"
	"math"
)

type T = float64
type TimeSeries = []T

/*
	FastDTW
	https://github.com/Citing/fastDTW
*/
func FastDTW(seriesX TimeSeries, seriesY TimeSeries, radius int) (distance T, path [][2]int) {
	var minTSsize = radius + 2
	if len(seriesX) < minTSsize || len(seriesY) < minTSsize {
		distance, path = DTW(seriesX, seriesY, nil)
	} else {
		var shrunkX = reduceByHalf(seriesX)
		var shrunkY = reduceByHalf(seriesY)
		var _, lowResPath = FastDTW(shrunkX, shrunkY, radius)
		var window = expandResWindow(lowResPath, len(seriesX), len(seriesY), radius)
		distance, path = DTW(seriesX, seriesY, window)
	}
	return
}

func DTW(seriesX TimeSeries, seriesY TimeSeries, window [][2]int) (T, [][2]int) {
	if window == nil {
		window = make([][2]int, (len(seriesX)+1)*(len(seriesY)+1))
		k := 0
		for i := 0; i <= len(seriesX); i++ {
			for j := 0; j <= len(seriesY); j++ {
				window[k] = [2]int{i, j}
				k++
			}
		}
	} else {
		for k := 0; k < len(window); k++ {
			window[k] = [2]int{window[k][0] + 1, window[k][1] + 1}
		}
	}
	type p struct {
		dist      T
		neighborX int
		neighborY int
	}
	var D = make(map[[2]int]p)
	for _, v := range window {
		// T type!
		D[v] = p{math.MaxFloat64, 0, 0}
	}
	D[[2]int{0, 0}] = p{0, 0, 0}
	for i := 1; i <= len(seriesX); i++ {
		for j := 1; j <= len(seriesY); j++ {
			v := [2]int{i, j}
			dt := math.Abs(seriesX[v[0]-1] - seriesY[v[1]-1])
			D[v] = func(p1 p, p2 p, p3 p) p {
				var tmp p
				if p1.dist < p2.dist {
					tmp = p1
				} else {
					tmp = p2
				}
				if p3.dist < tmp.dist {
					tmp = p3
				}
				return tmp
			}(p{D[[2]int{v[0] - 1, v[1]}].dist + dt, v[0] - 1, v[1]}, p{D[[2]int{v[0], v[1] - 1}].dist + dt, v[0], v[1] - 1}, p{D[[2]int{v[0] - 1, v[1] - 1}].dist + dt, v[0] - 1, v[1] - 1})
		}
	}

	var path_ list.List
	for i, j := len(seriesX), len(seriesY); i != 0 || j != 0; {
		path_.PushBack([2]int{i - 1, j - 1})
		i, j = D[[2]int{i, j}].neighborX, D[[2]int{i, j}].neighborY
	}
	path := make([][2]int, path_.Len())
	for i, j, k := len(seriesX), len(seriesY), 0; i != 0 || j != 0; k++ {
		path[k] = ([2]int{i - 1, j - 1})
		i, j = D[[2]int{i, j}].neighborX, D[[2]int{i, j}].neighborY
	}

	distance := D[[2]int{len(seriesX), len(seriesY)}].dist
	return distance, path
}

func reduceByHalf(series TimeSeries) TimeSeries {
	shrunk := make([]T, len(series)/2)
	for i := 0; i < len(shrunk); i++ {
		shrunk[i] = (series[2*i] + series[2*i+1]) / 2
	}
	return shrunk
}

func expandResWindow(path [][2]int, X int, Y int, radius int) [][2]int {
	var window_ = make(map[[2]int]int)

	for k := 0; k < len(path); k++ {
		for i := 0; i <= 2*radius; i++ {
			for j := 0; j <= 2*radius; j++ {
				x := 2 * (path[k][0] - radius + i)
				y := 2 * (path[k][1] - radius + j)
				window_[[2]int{x, y}] = 1
				window_[[2]int{x + 1, y}] = 1
				window_[[2]int{x, y + 1}] = 1
				window_[[2]int{x + 1, y + 1}] = 1
			}
		}
	}

	var window__ = make(map[[2]int]int)
	start_j := 0
	for i := 0; i < X; i++ {
		new_start_j := -2
		for j := start_j; j < Y; j++ {
			if window_[[2]int{i, j}] == 1 {
				window__[[2]int{i, j}] = 1
				if new_start_j == -2 {
					new_start_j = j
				}
			} else if new_start_j != -2 {
				break
			}
			start_j = new_start_j
		}
	}

	var window = make([][2]int, len(window__))
	start_j = 0
	k := 0
	for i := 0; i < X; i++ {
		new_start_j := -2
		for j := start_j; j < Y; j++ {
			if window_[[2]int{i, j}] == 1 {
				window[k] = [2]int{i, j}
				k++
				if new_start_j == -2 {
					new_start_j = j
				}
			} else if new_start_j != -2 {
				break
			}
			start_j = new_start_j
		}
	}
	return window
}
