package reactor

import (
	"errors"
	"fmt"

	"batch-arrh/internal/kinetics"
)

// Derivative is the right-hand side of an ODE in the conversion x.
type Derivative func(t, x float64) (float64, error)

// buildConversionDerivative assembles dX/dt for the constant-volume batch
// reactor. The material balance and the rate law share one chain:
//
//	(-r_A)*V = -dN_A/dt          material balance
//	N_A = C_A*V                  constant volume
//	C_A = C_A0*(1-X)             constant-volume conversion relation
//	=> dX/dt = (-r_A)/C_A0
//
// No variable-volume term appears anywhere: the scenario is declared
// isochoric and every concentration follows C_i = C_i0 + (nu_i/nu_A)*C_A0*X.
func (r *BatchReactor) buildConversionDerivative() (Derivative, error) {
	stoich := kinetics.NewStoich(r.cfg.Stoich)
	reactants := stoich.Reactants()
	if len(reactants) == 0 {
		return nil, kinetics.ErrNoReactant
	}
	nuA := stoich.Coeff(r.key)
	if nuA >= 0 {
		return nil, fmt.Errorf("key species %q has non-negative coefficient %g", r.key, nuA)
	}
	law := kinetics.NewLaw(r.order, r.k)
	cA0 := r.cfg.Initial[r.key]

	switch r.order {
	case kinetics.OrderFirst:
		// dX/dt = k*(1-X)
		return func(t, x float64) (float64, error) {
			cA := cA0 * (1 - x)
			rr, err := law.Rate(cA, 0)
			if err != nil {
				return 0, err
			}
			return rr / cA0, nil
		}, nil
	case kinetics.OrderSecond:
		// dX/dt = k*C_A*C_B/C_A0 with both concentrations derived from X
		// through the shared stoichiometry.
		other, err := r.secondReactant(reactants)
		if err != nil {
			return nil, err
		}
		nuB := stoich.Coeff(other)
		cB0 := r.cfg.Initial[other]
		return func(t, x float64) (float64, error) {
			cA := cA0 * (1 - x)
			cB := cB0 - (nuB/nuA)*cA0*x
			rr, err := law.Rate(cA, cB)
			if err != nil {
				return 0, err
			}
			return rr / cA0, nil
		}, nil
	default:
		return nil, kinetics.ErrUnknownOrder
	}
}

// secondReactant picks the co-reactant B of a bimolecular step. A second
// order law needs exactly two distinct reactants; anything else is a
// configuration error that must surface before integration starts.
func (r *BatchReactor) secondReactant(reactants []string) (string, error) {
	others := make([]string, 0, len(reactants))
	for _, name := range reactants {
		if name != r.key {
			others = append(others, name)
		}
	}
	if len(others) != 1 {
		return "", errors.New("second-order rate needs exactly two reactants")
	}
	return others[0], nil
}
