package kinetics

import "fmt"

// SecondOrder implements -r_A = k*C_A*C_B for a bimolecular step. When the
// two reactants share the same initial concentration and a 1:1
// stoichiometry the conversion equation reduces to 1/C_A - 1/C_A0 = k*t,
// which is the closed-form identity used by the cross-checks.
type SecondOrder struct {
	K float64
}

// Order identifies the law as second order.
func (s SecondOrder) Order() Order { return OrderSecond }

// Rate evaluates -r_A = k*C_A*C_B. Both concentrations must be
// non-negative.
func (s SecondOrder) Rate(cA, cB float64) (float64, error) {
	if cA < 0 {
		return 0, fmt.Errorf("%w: C_A = %g", ErrNegativeConcentration, cA)
	}
	if cB < 0 {
		return 0, fmt.Errorf("%w: C_B = %g", ErrNegativeConcentration, cB)
	}
	return s.K * cA * cB, nil
}

// HalfLife returns t_1/2 = 1/(k*C_A0) for a second-order reaction starting
// from a single concentration. It is only meaningful when a second reactant
// is present at the same concentration or is absent entirely.
func (s SecondOrder) HalfLife(cA0 float64) float64 {
	if s.K <= 0 || cA0 <= 0 {
		return 0
	}
	return 1 / (s.K * cA0)
}

// String renders the law for diagnostics.
func (s SecondOrder) String() string {
	return fmt.Sprintf("second-order, k=%.6g", s.K)
}
