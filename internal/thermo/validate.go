package thermo

import (
	"errors"
	"fmt"
)

// ErrNonPositiveTemperature means the absolute temperature is not strictly
// positive. The Arrhenius exponent blows up at T = 0, so such a value is
// never accepted.
var ErrNonPositiveTemperature = errors.New("temperature must be > 0 K")

// ErrNegativeActivationEnergy means Ea < 0, which has no physical meaning
// for an endothermic Arrhenius step.
var ErrNegativeActivationEnergy = errors.New("activation energy must be >= 0")

// ErrNonPositivePreExponential means A <= 0. A zero pre-exponential would
// silently turn every reaction off.
var ErrNonPositivePreExponential = errors.New("pre-exponential factor A must be > 0")

// ValidateTemperature rejects absolute temperatures that cannot feed the
// Arrhenius exponent.
func ValidateTemperature(t float64) error {
	if t <= 0 {
		return ErrNonPositiveTemperature
	}
	return nil
}

// ValidateActivationEnergy rejects negative activation energies.
func ValidateActivationEnergy(ea float64) error {
	if ea < 0 {
		return ErrNegativeActivationEnergy
	}
	return nil
}

// ValidatePreExponential rejects non-positive pre-exponential factors.
func ValidatePreExponential(a float64) error {
	if a <= 0 {
		return ErrNonPositivePreExponential
	}
	return nil
}

// ValidateArrhenius runs all component checks of an Arrhenius triplet.
func ValidateArrhenius(p Arrhenius) error {
	if err := ValidateTemperature(p.T); err != nil {
		return err
	}
	if err := ValidateActivationEnergy(p.Ea); err != nil {
		return err
	}
	if err := ValidatePreExponential(p.A); err != nil {
		return err
	}
	return nil
}

// Describe assembles a one-line, human readable rendering of the rate
// parameters, used by the text report.
func Describe(p Arrhenius) string {
	return fmt.Sprintf("A=%.4g, Ea=%.4g J/mol, T=%.4g K", p.A, p.Ea, p.T)
}
