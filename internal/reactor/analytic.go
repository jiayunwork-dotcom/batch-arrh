package reactor

import "math"

func FirstOrderConversion(k, t float64) float64 {
	if k < 0 || t < 0 {
		return math.NaN()
	}
	return 1 - math.Exp(-k*t)
}

func FirstOrderConcentration(cA0, k, t float64) float64 {
	if k < 0 || t < 0 || cA0 < 0 {
		return math.NaN()
	}
	return cA0 * math.Exp(-k*t)
}

func SecondOrderEqualConversion(k, cA0, t float64) float64 {
	if k < 0 || t < 0 || cA0 <= 0 {
		return math.NaN()
	}
	return k * cA0 * t / (1 + k*cA0*t)
}

func SecondOrderEqualConcentration(cA0, k, t float64) float64 {
	if k < 0 || t < 0 || cA0 <= 0 {
		return math.NaN()
	}
	return cA0 / (1 + k*cA0*t)
}

func InverseConcentrationDeficit(cA, cA0 float64) float64 {
	if cA <= 0 || cA0 <= 0 {
		return math.NaN()
	}
	return 1/cA - 1/cA0
}

func ConversionRemainder(x float64) float64 {
	return 1 - x
}
