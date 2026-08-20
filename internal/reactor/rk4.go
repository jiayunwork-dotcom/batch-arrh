package reactor

import (
	"math"
)

// rk4 integrates the conversion ODE dx/dt = f(t, x) over n fixed-size
// steps from t0 to t1 with the classical fourth-order Runge-Kutta scheme.
// The returned slice has n+1 entries starting at x(t0). Local truncation
// error per step is O(h^5) and the global error O(h^4), which keeps the
// second-order integration in tight agreement with the closed form.
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

// timeGrid builds the time stamps aligned with an n-step integration from
// t0 to t1. The grid always starts exactly at t0 so a zero-time run
// trivially returns the initial state.
func timeGrid(t0, t1 float64, n int) []float64 {
	h := (t1 - t0) / float64(n)
	ts := make([]float64, n+1)
	for i := range ts {
		ts[i] = t0 + float64(i)*h
	}
	return ts
}

// clampConversion keeps a numerically integrated conversion inside [0,1].
// RK4 on a bounded first-order field can overshoot by an epsilon near the
// asymptote; clamping only corrects that numerical noise and never changes
// a genuinely interior value.
func clampConversion(x float64) float64 {
	return math.Max(0, math.Min(1, x))
}
