package kinetics

import (
	"errors"
	"fmt"
	"math"
)

var ErrNegativeConcentration = errors.New("negative concentration")

var ErrNegativeRateConstant = errors.New("rate constant must be >= 0")

var ErrNoReactant = errors.New("stoichiometry has no reactant species")

var ErrUnknownOrder = errors.New("unknown rate order")

func CheckConcentration(name string, c float64) error {
	if c < 0 {
		return fmt.Errorf("%w: %s = %g", ErrNegativeConcentration, name, c)
	}
	return nil
}

func CheckRateConstant(k float64) error {
	if k < 0 {
		return fmt.Errorf("%w: %g", ErrNegativeRateConstant, k)
	}
	return nil
}

func NearlyEqual(a, b, relTol float64) bool {
	if math.Abs(a-b) <= relTol {
		return true
	}
	return math.Abs(a-b) <= relTol*math.Max(math.Abs(a), math.Abs(b))
}
