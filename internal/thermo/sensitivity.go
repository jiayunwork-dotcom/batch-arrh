package thermo

import (
	"errors"
	"math"
)

func DlnKdT(ea, t float64) (float64, error) {
	if err := ValidateTemperature(t); err != nil {
		return 0, err
	}
	if err := ValidateActivationEnergy(ea); err != nil {
		return 0, err
	}
	return ea / (R * t * t), nil
}

func TemperatureToScaleRate(p Arrhenius, factor float64) (float64, error) {
	if factor <= 0 {
		return 0, ErrNonPositiveRateConstant
	}
	if err := ValidateArrhenius(p); err != nil {
		return 0, err
	}
	if p.Ea == 0 {
		if factor == 1 {
			return p.T, nil
		}
		return 0, errors.New("zero activation energy cannot scale k by changing T")
	}
	k, err := p.Rate()
	if err != nil {
		return 0, err
	}
	return p.TemperatureForRate(k * factor)
}

func Q10(p Arrhenius) (float64, error) {
	k1, err := p.Rate()
	if err != nil {
		return 0, err
	}
	k2, err := p.RateAt(p.T + 10)
	if err != nil {
		return 0, err
	}
	return k2 / k1, nil
}

func FiniteDifferenceDlnKdT(p Arrhenius, dt float64) (float64, error) {
	if dt <= 0 {
		return 0, errors.New("finite-difference dt must be > 0")
	}
	k0, err := p.Rate()
	if err != nil {
		return 0, err
	}
	k1, err := p.RateAt(p.T + dt)
	if err != nil {
		return 0, err
	}
	if k0 <= 0 {
		return 0, ErrNonPositiveRateConstant
	}
	return math.Log(k1/k0) / dt, nil
}

func TwoPointSlope(k1, t1, k2, t2 float64) (float64, error) {
	if k1 <= 0 || k2 <= 0 {
		return 0, ErrNonPositiveRateConstant
	}
	if err := ValidateTemperature(t1); err != nil {
		return 0, err
	}
	if err := ValidateTemperature(t2); err != nil {
		return 0, err
	}
	if t1 == t2 {
		return 0, errors.New("the two temperatures coincide")
	}
	return math.Log(k2/k1) / (1/t2 - 1/t1), nil
}
