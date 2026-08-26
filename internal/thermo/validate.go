package thermo

import (
	"errors"
	"fmt"
)

var ErrNonPositiveTemperature = errors.New("temperature must be > 0 K")

var ErrNegativeActivationEnergy = errors.New("activation energy must be >= 0")

var ErrNonPositivePreExponential = errors.New("pre-exponential factor A must be > 0")

func ValidateTemperature(t float64) error {
	if t <= 0 {
		return ErrNonPositiveTemperature
	}
	return nil
}

func ValidateActivationEnergy(ea float64) error {
	if ea < 0 {
		return ErrNegativeActivationEnergy
	}
	return nil
}

func ValidatePreExponential(a float64) error {
	if a <= 0 {
		return ErrNonPositivePreExponential
	}
	return nil
}

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

func Describe(p Arrhenius) string {
	return fmt.Sprintf("A=%.4g, Ea=%.4g J/mol, T=%.4g K", p.A, p.Ea, p.T)
}
