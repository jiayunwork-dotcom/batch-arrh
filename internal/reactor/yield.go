package reactor

import "math"

func IntermediateYield(cB, cA0 float64) float64 {
	if cA0 <= 0 {
		return math.NaN()
	}
	return cB / cA0
}

func OverallSelectivity(cB, cA, cA0 float64) float64 {
	consumed := cA0 - cA
	if consumed <= 0 {
		return 0
	}
	return cB / consumed
}

func YieldFromSelectivity(s, x float64) float64 {
	return s * x
}

func InstantaneousSeriesSelectivity(k1, k2, cA, cB float64) float64 {
	if k1 <= 0 || cA <= 0 {
		return math.NaN()
	}
	rA := k1 * cA
	rB := k1*cA - k2*cB
	if rA == 0 {
		return math.NaN()
	}
	return rB / rA
}

func SeriesPeakYield(cA0, k1, k2 float64) float64 {
	cB := SeriesPeakConcentration(cA0, k1, k2)
	return IntermediateYield(cB, cA0)
}
