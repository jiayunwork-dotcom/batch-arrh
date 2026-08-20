package kinetics

import (
	"errors"
	"fmt"
	"math"
)

// ErrNegativeConcentration is returned by rate laws when a concentration
// argument is negative; negative concentrations are rejected by validation
// upstream, so this error only fires on programmer misuse.
var ErrNegativeConcentration = errors.New("negative concentration")

// ErrNegativeRateConstant is returned when a scenario supplies k < 0, which
// would integrate backwards in time.
var ErrNegativeRateConstant = errors.New("rate constant must be >= 0")

// ErrNoReactant is returned when a stoichiometry has no species on the
// reactant side, leaving the conversion key undefined.
var ErrNoReactant = errors.New("stoichiometry has no reactant species")

// ErrUnknownOrder is returned for a rate order string the solver does not
// implement.
var ErrUnknownOrder = errors.New("unknown rate order")

// CheckConcentration validates a single concentration value against the
// physical lower bound.
func CheckConcentration(name string, c float64) error {
	if c < 0 {
		return fmt.Errorf("%w: %s = %g", ErrNegativeConcentration, name, c)
	}
	return nil
}

// CheckRateConstant validates a direct rate constant.
func CheckRateConstant(k float64) error {
	if k < 0 {
		return fmt.Errorf("%w: %g", ErrNegativeRateConstant, k)
	}
	return nil
}

// NearlyEqual compares two floats with a relative tolerance, used by the
// material-balance cross-checks where both sides of an equation come from
// independent code paths.
func NearlyEqual(a, b, relTol float64) bool {
	if math.Abs(a-b) <= relTol {
		return true
	}
	return math.Abs(a-b) <= relTol*math.Max(math.Abs(a), math.Abs(b))
}
