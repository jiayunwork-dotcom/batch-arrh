package reactor

import (
	"errors"
	"math"
)

type SeriesPoint struct {
	Time        float64
	CA          float64
	CB          float64
	CC          float64
	Selectivity float64
}

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

func SeriesPeakTime(k1, k2 float64) float64 {
	if k1 <= 0 || k2 <= 0 {
		return math.NaN()
	}
	if math.Abs(k2-k1) < 1e-15 {
		return 1 / k1
	}
	return math.Log(k1/k2) / (k1 - k2)
}

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

func FillSeriesConcentrations(p SeriesPoint, dst map[string]float64) map[string]float64 {
	if dst == nil {
		dst = make(map[string]float64, 3)
	}
	dst["A"] = p.CA
	dst["B"] = p.CB
	dst["C"] = p.CC
	return dst
}

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
