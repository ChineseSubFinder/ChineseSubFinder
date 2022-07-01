package dtw

import (
	"testing"
)

func TestFastDTW(t *testing.T) {
	var balance1 = TimeSeries{1, 2, 3, 4, 5}
	var balance2 = TimeSeries{1, 1.3, 2, 3, 5, 6}
	//var balance2 = TimeSeries{1, 2, 3.1, 4, 5}
	d, b := FastDTW(balance1, balance2, 1)

	t.Logf("\n\nd: %f\nb: %d", d, b)

}