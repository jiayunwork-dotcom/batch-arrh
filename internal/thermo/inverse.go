package thermo

import (
	"errors"
	"math"
)

var ErrNonPositiveRateConstant = errors.New("rate constant must be > 0")

func ActivationEnergyFromPoints(k1, t1, k2, t2 float64) (float64, error) {
	if k1 <= 0 || k2 <= 0 {
		return 0, ErrNonPositiveRateConstant
	}
	if err := ValidateTemperature(t1); err != nil {
		return 0, err
	}
	if err := ValidateTemperature(t2); err != nil {
		return 0, err
	}
	invDiff := 1/t1 - 1/t2
	if invDiff == 0 {
		return 0, errors.New("the two temperatures coincide")
	}
	return R * math.Log(k2/k1) / invDiff, nil
}

func PreExponentialFromPoints(k, t, ea float64) (float64, error) {
	if k <= 0 {
		return 0, ErrNonPositiveRateConstant
	}
	if err := ValidateTemperature(t); err != nil {
		return 0, err
	}
	if err := ValidateActivationEnergy(ea); err != nil {
		return 0, err
	}
	return k * math.Exp(ea/(R*t)), nil
}

func (p Arrhenius) TemperatureForRate(kTarget float64) (float64, error) {
	if kTarget <= 0 {
		return 0, ErrNonPositiveRateConstant
	}
	if err := ValidateArrhenius(p); err != nil {
		return 0, err
	}
	if kTarget >= p.A {
		return math.Inf(1), nil
	}
	return p.Ea / (R * math.Log(p.A/kTarget)), nil
}

func (p Arrhenius) VerifySelfConsistent(k float64, relTol float64) (bool, error) {
	computed, err := p.Rate()
	if err != nil {
		return false, err
	}
	if math.Abs(computed) < relTol {
		return k == computed, nil
	}
	return math.Abs(k-computed) <= relTol*math.Abs(computed), nil
}
