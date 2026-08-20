package reactor

import (
	"errors"
	"math"
)

// SeriesPoint is one snapshot of the A -> B -> C network at time t. The
// molar balance C_A + C_B + C_C = C_A0 holds at every snapshot because C_C
// is always derived from the difference; species B never gains or loses
// atoms, only moles are redistributed.
type SeriesPoint struct {
	Time        float64
	CA          float64
	CB          float64
	CC          float64
	Selectivity float64
}

// SolveSeries integrates the consecutive first-order network A -> B -> C
// with the explicit analytic solution:
//
//	C_A = C_A0*exp(-k1*t)
//	C_B = C_A0*k1/(k2-k1)*(exp(-k1*t) - exp(-k2*t))
//	C_C = C_A0 - C_A - C_B     (molar conservation, pinned)
//
// The special case k1 == k2 uses the degenerate form C_B = C_A0*k*t*e^(-k*t).
func SolveSeries(cA0, k1, k2, t float64, steps int) ([]SeriesPoint, error) {
	if cA0 <= 0 {
		return nil, errors.New("series reaction needs C_A0 > 0")
	}
	if k1 <= 0 || k2 <= 0 {
		return nil, ErrNonPositiveSeriesConstants
	}
	if t < 0 {
		return nil, ErrNegativeTime
	}
	if steps < 1 {
		steps = 2000
	}
	ts := timeGrid(0, t, steps)
	pts := make([]SeriesPoint, len(ts))
	for i, tt := range ts {
		cA := cA0 * math.Exp(-k1*tt)
		var cB float64
		if math.Abs(k2-k1) < 1e-15 {
			cB = cA0 * k1 * tt * math.Exp(-k1*tt)
		} else {
			cB = cA0 * k1 / (k2 - k1) * (math.Exp(-k1*tt) - math.Exp(-k2*tt))
		}
		cC := cA0 - cA - cB
		consumed := cA0 - cA
		var sel float64
		if consumed > 0 {
			sel = cB / consumed
		}
		pts[i] = SeriesPoint{Time: tt, CA: cA, CB: cB, CC: cC, Selectivity: sel}
	}
	return pts, nil
}

// SeriesPeakTime returns the residence time at which the intermediate B
// reaches its maximum concentration: t_max = ln(k1/k2)/(k1-k2) for
// k1 != k2, and t_max = 1/k for the degenerate k1 == k2 case.
func SeriesPeakTime(k1, k2 float64) float64 {
	if k1 <= 0 || k2 <= 0 {
		return math.NaN()
	}
	if math.Abs(k2-k1) < 1e-15 {
		return 1 / k1
	}
	return math.Log(k1/k2) / (k1 - k2)
}

// SeriesPeakConcentration returns C_B at the peak time, the analytic value
// against which the trajectory maximum is compared.
func SeriesPeakConcentration(cA0, k1, k2 float64) float64 {
	tp := SeriesPeakTime(k1, k2)
	if math.IsNaN(tp) {
		return math.NaN()
	}
	if math.Abs(k2-k1) < 1e-15 {
		return cA0 * k1 * tp * math.Exp(-k1*tp)
	}
	return cA0 * k1 / (k2 - k1) * (math.Exp(-k1*tp) - math.Exp(-k2*tp))
}

// PeakTimeFromTrajectory scans a series trajectory and returns the time of
// the snapshot with the largest C_B. It is used to confirm that the
// numerical trajectory places the peak at the analytic t_max within one
// integration step.
func PeakTimeFromTrajectory(pts []SeriesPoint) float64 {
	if len(pts) == 0 {
		return math.NaN()
	}
	best := pts[0]
	for _, p := range pts[1:] {
		if p.CB > best.CB {
			best = p
		}
	}
	return best.Time
}
