package kinetics

import (
	"fmt"
	"sort"
	"strings"
)

type Stoich struct {
	Coeffs map[string]float64
}

func NewStoich(coeffs map[string]float64) Stoich {
	cp := make(map[string]float64, len(coeffs))
	for name, c := range coeffs {
		cp[name] = c
	}
	return Stoich{Coeffs: cp}
}

func (s Stoich) Coeff(name string) float64 {
	return s.Coeffs[name]
}

func (s Stoich) Species() []string {
	names := make([]string, 0, len(s.Coeffs))
	for name, c := range s.Coeffs {
		if c != 0 {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func (s Stoich) Reactants() []string {
	out := []string{}
	for _, name := range s.Species() {
		if s.Coeffs[name] < 0 {
			out = append(out, name)
		}
	}
	return out
}

func (s Stoich) Products() []string {
	out := []string{}
	for _, name := range s.Species() {
		if s.Coeffs[name] > 0 {
			out = append(out, name)
		}
	}
	return out
}

func (s Stoich) HasReactant(name string) bool {
	return s.Coeffs[name] < 0
}

func (s Stoich) HasProduct(name string) bool {
	return s.Coeffs[name] > 0
}

func (s Stoich) ConsumptionRatio(a, b string) (float64, error) {
	if !s.HasReactant(a) {
		return 0, fmt.Errorf("%w: %q is not a reactant", ErrNoReactant, a)
	}
	if !s.HasReactant(b) {
		return 0, fmt.Errorf("%w: %q is not a reactant", ErrNoReactant, b)
	}
	return s.Coeffs[b] / s.Coeffs[a], nil
}

func (s Stoich) ProductionRatio(product, reactant string) (float64, error) {
	if !s.HasProduct(product) {
		return 0, fmt.Errorf("%w: %q is not a product", ErrNoReactant, product)
	}
	if !s.HasReactant(reactant) {
		return 0, fmt.Errorf("%w: %q is not a reactant", ErrNoReactant, reactant)
	}
	return s.Coeffs[product] / -s.Coeffs[reactant], nil
}

func (s Stoich) Validate() error {
	if len(s.Reactants()) == 0 {
		return ErrNoReactant
	}
	if len(s.Products()) == 0 {
		return fmt.Errorf("stoichiometry has no product species")
	}
	return nil
}

func (s Stoich) String() string {
	var lhs, rhs []string
	for _, name := range s.Species() {
		c := s.Coeffs[name]
		term := name
		if mathAbs(c) != 1 {
			term = fmt.Sprintf("%g%s", mathAbs(c), name)
		}
		if c < 0 {
			lhs = append(lhs, term)
		} else {
			rhs = append(rhs, term)
		}
	}
	return strings.Join(lhs, " + ") + " -> " + strings.Join(rhs, " + ")
}

func mathAbs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
