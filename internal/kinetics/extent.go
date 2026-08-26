package kinetics

import "math"

func ExtentFromConversion(cA0, vol, x, nuA float64) float64 {
	if nuA >= 0 || vol <= 0 || cA0 < 0 || x < 0 {
		return math.NaN()
	}
	return vol * cA0 * x / (-nuA)
}

func MolesFromExtent(n0, nu, xi float64) float64 {
	return n0 + nu*xi
}

func ConversionFromExtent(xi, cA0, vol, nuA float64) float64 {
	if cA0 <= 0 || vol <= 0 || nuA >= 0 {
		return math.NaN()
	}
	return xi * (-nuA) / (cA0 * vol)
}

func FractionalConversion(cA, cA0 float64) float64 {
	if cA0 <= 0 {
		return math.NaN()
	}
	return 1 - cA/cA0
}

func RemainingFraction(x float64) float64 {
	return 1 - x
}
