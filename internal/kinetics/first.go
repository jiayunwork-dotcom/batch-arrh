package kinetics

import "fmt"

type FirstOrder struct {
	K float64
}

func (f FirstOrder) Order() Order { return OrderFirst }

func (f FirstOrder) Rate(cA, _ float64) (float64, error) {
	if cA < 0 {
		return 0, fmt.Errorf("%w: C_A = %g", ErrNegativeConcentration, cA)
	}
	return f.K * cA, nil
}

func (f FirstOrder) HalfLife() float64 {
	if f.K <= 0 {
		return 0
	}
	return ln2 / f.K
}

func (f FirstOrder) String() string {
	return fmt.Sprintf("first-order, k=%.6g", f.K)
}

const ln2 = 0.6931471805599453
