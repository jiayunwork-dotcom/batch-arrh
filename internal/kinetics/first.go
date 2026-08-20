package kinetics

import "fmt"

// FirstOrder implements -r_A = k*C_A for the key reactant A. The rate law
// is linear in C_A, which makes the conversion equation integrable in
// closed form: X = 1 - exp(-k*t).
type FirstOrder struct {
	K float64
}

// Order identifies the law as first order.
func (f FirstOrder) Order() Order { return OrderFirst }

// Rate evaluates -r_A = k*C_A. The co-reactant argument is ignored by the
// first-order law; it is accepted so every Law shares one call signature.
func (f FirstOrder) Rate(cA, _ float64) (float64, error) {
	if cA < 0 {
		return 0, fmt.Errorf("%w: C_A = %g", ErrNegativeConcentration, cA)
	}
	return f.K * cA, nil
}

// HalfLife returns t_1/2 = ln(2)/k for a first-order reaction. It is a
// natural companion of the rate law and is used by the report to give the
// operator a feel for the time scale.
func (f FirstOrder) HalfLife() float64 {
	if f.K <= 0 {
		return 0
	}
	return ln2 / f.K
}

// String renders the law for diagnostics.
func (f FirstOrder) String() string {
	return fmt.Sprintf("first-order, k=%.6g", f.K)
}

// ln2 is the natural logarithm of two, kept local to avoid importing math
// just for one constant at every call site.
const ln2 = 0.6931471805599453
