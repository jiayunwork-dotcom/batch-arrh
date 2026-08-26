package reactor

import (
	"math"
)

func rk4(f Derivative, x0, t0, t1 float64, n int) ([]float64, error) {
	h := (t1 - t0) / float64(n)
	xs := make([]float64, n+1)
	xs[0] = x0
	t := t0
	x := x0
	for i := 0; i < n; i++ {
		k1, err := f(t, x)
		if err != nil {
			return nil, err
		}
		k2, err := f(t+h/2, x+h*k1/2)
		if err != nil {
			return nil, err
		}
		k3, err := f(t+h/2, x+h*k2/2)
		if err != nil {
			return nil, err
		}
		k4, err := f(t+h, x+h*k3)
		if err != nil {
			return nil, err
		}
		x += h / 6 * (k1 + 2*k2 + 2*k3 + k4)
		t += h
		xs[i+1] = x
	}
	return xs, nil
}

func timeGrid(t0, t1 float64, n int) []float64 {
	h := (t1 - t0) / float64(n)
	ts := make([]float64, n+1)
	for i := range ts {
		ts[i] = t0 + float64(i)*h
	}
	return ts
}

func clampConversion(x float64) float64 {
	return math.Max(0, math.Min(1, x))
}
