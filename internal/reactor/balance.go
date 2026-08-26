package reactor

import (
	"errors"
	"fmt"

	"batch-arrh/internal/kinetics"
)

type Derivative func(t, x float64) (float64, error)

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
		return func(t, x float64) (float64, error) {
			cA := cA0 * (1 - x)
			rr, err := law.Rate(cA, 0)
			if err != nil {
				return 0, err
			}
			return rr / cA0, nil
		}, nil
	case kinetics.OrderSecond:
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
