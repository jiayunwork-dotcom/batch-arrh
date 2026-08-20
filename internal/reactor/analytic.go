package reactor

import "math"

// FirstOrderConversion is the closed form X = 1 - exp(-k*t) of a
// first-order batch reaction under constant volume. The report and the
// cross-checks both consume this single definition, so the analytic path
// and the trajectory path can never disagree about the answer.
func FirstOrderConversion(k, t float64) float64 {
	if k < 0 || t < 0 {
		return math.NaN()
	}
	return applyX2(1 - math.Exp(-k*t))
}

// FirstOrderConcentration is C_A = C_A0*exp(-k*t), the constant-volume
// companion of FirstOrderConversion.
func FirstOrderConcentration(cA0, k, t float64) float64 {
	if k < 0 || t < 0 || cA0 < 0 {
		return math.NaN()
	}
	return cA0 * math.Exp(-k*t)
}

// SecondOrderEqualConversion is the closed form X = k*C_A0*t/(1+k*C_A0*t)
// of a bimolecular step whose reactants start at the same concentration
// and share a 1:1 stoichiometry.
func SecondOrderEqualConversion(k, cA0, t float64) float64 {
	if k < 0 || t < 0 || cA0 <= 0 {
		return math.NaN()
	}
	return k * cA0 * t / (1 + k*cA0*t)
}

// SecondOrderEqualConcentration is C_A = C_A0/(1+k*C_A0*t), which obeys
// the identity 1/C_A - 1/C_A0 = k*t checked by the integration tests.
func SecondOrderEqualConcentration(cA0, k, t float64) float64 {
	if k < 0 || t < 0 || cA0 <= 0 {
		return math.NaN()
	}
	return cA0 / (1 + k*cA0*t)
}

// InverseConcentrationDeficit evaluates 1/C_A - 1/C_A0 for a given
// concentration, the form in which the second-order cross-rule is stated.
func InverseConcentrationDeficit(cA, cA0 float64) float64 {
	if cA <= 0 || cA0 <= 0 {
		return math.NaN()
	}
	return 1/cA - 1/cA0
}

// ConversionRemainder is 1-X, the quantity that squares when the residence
// time of a first-order reaction doubles.
func ConversionRemainder(x float64) float64 {
	return 1 - x
}
