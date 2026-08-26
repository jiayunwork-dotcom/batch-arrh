package reactor

import "math"

func SecondOrderUnequalConversion(k, cA0, cB0, t float64) float64 {
	if k < 0 || t < 0 || cA0 <= 0 || cB0 <= 0 {
		return math.NaN()
	}
	m := cB0 / cA0
	if math.Abs(m-1) < 1e-12 {
		return SecondOrderEqualConversion(k, cA0, t)
	}
	theta := math.Exp((1 - m) * cA0 * k * t)
	den := m - theta
	if den == 0 {
		return math.NaN()
	}
	return m * (1 - theta) / den
}

func SecondOrderUnequalConcentrationA(cA0, k, cB0, t float64) float64 {
	x := SecondOrderUnequalConversion(k, cA0, cB0, t)
	if math.IsNaN(x) {
		return math.NaN()
	}
	return cA0 * (1 - x)
}

func SecondOrderUnequalConcentrationB(cA0, k, cB0, t float64) float64 {
	x := SecondOrderUnequalConversion(k, cA0, cB0, t)
	if math.IsNaN(x) {
		return math.NaN()
	}
	return cB0 - cA0*x
}

func LimitingConversion(cA0, cB0, nuA, nuB float64) float64 {
	if cA0 <= 0 || nuA >= 0 {
		return math.NaN()
	}
	if nuB >= 0 {
		return 1
	}
	ratio := (nuB / nuA) * (cA0 / cB0)
	if ratio <= 0 {
		return math.NaN()
	}
	if cB0/cA0 >= -nuB/(-nuA) {
		return 1
	}
	return cB0 / cA0 * (-nuA / -nuB)
}
