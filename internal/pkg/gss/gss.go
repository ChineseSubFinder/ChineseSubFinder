package gss

import (
	"log"
	"math"
)

var (
	sqrt5   = math.Sqrt(5)
	invphi  = (sqrt5 - 1) / 2 //# 1/phi
	invphi2 = (3 - sqrt5) / 2 //# 1/phi^2
	nan     = math.NaN()
)

// Gss golden section search (recursive version)
// https://en.wikipedia.org/wiki/Golden-section_search
// https://github.com/pa-m/optimize/blob/master/gss.go
// '''
// Golden section search, recursive.
// Given a function f with a single local minimum in
// the interval [a,b], gss returns a subset interval
// [c,d] that contains the minimum with d-c <= tol.
//
// logger may be nil
//
// example:
// >>> f = lambda x: (x-2)**2
// >>> a = 1
// >>> b = 5
// >>> tol = 1e-5
// >>> (c,d) = gssrec(f, a, b, tol)
// >>> print (c,d)
// (1.9999959837979107, 2.0000050911830893)
// '''
func Gss(f func(float64) float64, a, b, tol float64, logger *log.Logger) (float64, float64) {
	return gss(f, a, b, tol, nan, nan, nan, nan, nan, logger)
}
func gss(f func(float64) float64, a, b, tol, h, c, d, fc, fd float64, logger *log.Logger) (float64, float64) {
	if a > b {
		a, b = b, a
	}
	h = b - a
	it := 0
	for {
		if logger != nil {
			logger.Printf("%d\t%9.6g\t%9.6g\n", it, a, b)
		}
		it++
		if h < tol {
			return a, b
		}
		if a > b {
			a, b = b, a
		}
		if math.IsNaN(c) {
			c = a + invphi2*h
			fc = f(c)
		}
		if math.IsNaN(d) {
			d = a + invphi*h
			fd = f(d)
		}
		if fc < fd {
			b, h, c, fc, d, fd = d, h*invphi, nan, nan, c, fc
		} else {
			a, h, c, fc, d, fd = c, h*invphi, d, fd, nan, nan
		}
	}
}
