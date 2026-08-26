package reactor

import (
	"errors"
	"math"
)

func FirstOrderDamkohler(k, t float64) float64 {
	if k < 0 || t < 0 {
		return math.NaN()
	}
	return k * t
}

func ConversionFromFirstOrderDa(da float64) float64 {
	if da < 0 {
		return math.NaN()
	}
	return 1 - math.Exp(-da)
}

func SecondOrderEqualDamkohler(k, cA0, t float64) float64 {
	if k < 0 || cA0 <= 0 || t < 0 {
		return math.NaN()
	}
	return k * cA0 * t
}

func ConversionFromSecondOrderEqualDa(da float64) float64 {
	if da < 0 {
		return math.NaN()
	}
	return da / (1 + da)
}

func FirstOrderTimeForConversion(k, x float64) (float64, error) {
	if k <= 0 {
		return 0, errors.New("first-order design needs k > 0")
	}
	if x < 0 || x >= 1 {
		return 0, errors.New("target conversion must lie in [0, 1)")
	}
	if x == 0 {
		return 0, nil
	}
	return -math.Log(1-x) / k, nil
}

func SecondOrderEqualTimeForConversion(k, cA0, x float64) (float64, error) {
	if k <= 0 || cA0 <= 0 {
		return 0, errors.New("second-order design needs k > 0 and C_A0 > 0")
	}
	if x < 0 || x >= 1 {
		return 0, errors.New("target conversion must lie in [0, 1)")
	}
	if x == 0 {
		return 0, nil
	}
	return x / (k * cA0 * (1 - x)), nil
}

func HalfLifeResidenceRatio(t, tHalf float64) float64 {
	if tHalf <= 0 {
		return math.NaN()
	}
	return t / tHalf
}
