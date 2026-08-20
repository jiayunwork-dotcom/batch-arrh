package thermo

import (
	"errors"
	"math"
)

// ErrNonPositiveRateConstant is returned by the inverse Arrhenius helpers
// when a measured rate constant is not positive.
var ErrNonPositiveRateConstant = errors.New("rate constant must be > 0")

// ActivationEnergyFromPoints recovers the activation energy from two
// measured rate constants at two temperatures, the standard two-point
// Arrhenius fit:
//
//	Ea = R * ln(k2/k1) / (1/T1 - 1/T2)
//
// Both temperatures must be positive and both rate constants strictly
// positive; any violation is an error.
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

// PreExponentialFromPoints recovers the pre-exponential factor A from one
// measured rate constant, a temperature and a known activation energy:
// A = k * exp(Ea/(R*T)).
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

// TemperatureForRate inverts the Arrhenius law: it returns the absolute
// temperature at which the triplet produces a target rate constant,
// T = Ea/(R*ln(A/k)). When the target reaches A the required temperature
// is infinite, since the law saturates at A as T -> inf.
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

// VerifySelfConsistent checks that a measured rate constant agrees with the
// Arrhenius triplet within a relative tolerance. It is the guard used by
// the cross-check tests to prove k and the triplet describe one rate.
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
