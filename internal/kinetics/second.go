package kinetics

import "fmt"

type SecondOrder struct {
	K float64
}

func (s SecondOrder) Order() Order { return OrderSecond }

func (s SecondOrder) Rate(cA, cB float64) (float64, error) {
	if cA < 0 {
		return 0, fmt.Errorf("%w: C_A = %g", ErrNegativeConcentration, cA)
	}
	if cB < 0 {
		return 0, fmt.Errorf("%w: C_B = %g", ErrNegativeConcentration, cB)
	}
	return s.K * cA * cB, nil
}

func (s SecondOrder) HalfLife(cA0 float64) float64 {
	if s.K <= 0 || cA0 <= 0 {
		return 0
	}
	return 1 / (s.K * cA0)
}

func (s SecondOrder) String() string {
	return fmt.Sprintf("second-order, k=%.6g", s.K)
}
