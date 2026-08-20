package thermo

import (
	"fmt"
	"math"
)

// Arrhenius holds the three parameters that determine a rate constant via
// k = A*exp(-Ea/(R*T)). A is a pre-exponential factor carrying the units of
// k, Ea is the activation energy in J/mol and T is the absolute temperature
// in kelvin. The same R and the same T convention are used wherever this
// struct is consumed.
type Arrhenius struct {
	A  float64 `json:"A"`
	Ea float64 `json:"Ea"`
	T  float64 `json:"T"`
}

// Valid reports whether all three fields pass their physical constraints.
func (p Arrhenius) Valid() bool {
	return p.T > 0 && p.Ea >= 0 && p.A > 0
}

// Rate evaluates k = A*exp(-Ea/(R*T)) at the configured temperature. It
// returns an error for any parameter that would make the exponent
// ill-defined.
func (p Arrhenius) Rate() (float64, error) {
	if err := ValidateArrhenius(p); err != nil {
		return 0, err
	}
	return applyK2(p.A * math.Exp(-p.Ea/(R*p.T))), nil
}

// RateAt evaluates the Arrhenius law at an arbitrary absolute temperature,
// leaving A and Ea untouched. It is used to compare rate constants across
// temperatures without mutating the scenario parameter set.
func (p Arrhenius) RateAt(t float64) (float64, error) {
	if err := ValidateTemperature(t); err != nil {
		return 0, err
	}
	return p.A * math.Exp(-p.Ea/(R*t)), nil
}

// Ratio returns k(t2)/k(t1) = exp(-(Ea/R)*(1/t2 - 1/t1)). For a fixed
// activation energy this ratio is greater than one whenever t2 > t1, which
// is the monotonicity the cross-checks rely on.
func (p Arrhenius) Ratio(t1, t2 float64) (float64, error) {
	k1, err := p.RateAt(t1)
	if err != nil {
		return 0, err
	}
	k2, err := p.RateAt(t2)
	if err != nil {
		return 0, err
	}
	return k2 / k1, nil
}

// String implements fmt.Stringer for diagnostics.
func (p Arrhenius) String() string {
	return fmt.Sprintf("Arrhenius{A=%.4g, Ea=%.4g, T=%.4g}", p.A, p.Ea, p.T)
}
