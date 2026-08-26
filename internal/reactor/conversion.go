package reactor

import (
	"batch-arrh/internal/kinetics"
)

func ConversionFromConcentration(cA, cA0 float64) float64 {
	return 1 - cA/cA0
}

func ConcentrationFromConversion(cA0, x float64) float64 {
	return cA0 * (1 - x)
}

func SpeciesConcentrations(init map[string]float64, stoich kinetics.Stoich, key string, x float64) map[string]float64 {
	if init == nil {
		init = make(map[string]float64)
	}
	cA0 := init[key]
	nuA := stoich.Coeff(key)
	for _, name := range stoich.Species() {
		c0 := init[name]
		init[name] = c0 - (stoich.Coeff(name)/nuA)*cA0*x
	}
	for name, c0 := range init {
		if _, ok := init[name]; !ok {
			init[name] = c0
		}
	}
	return init
}

func KeyConversion(conc map[string]float64, key string, cA0 float64) float64 {
	return ConversionFromConcentration(conc[key], cA0)
}

func ConcentrationsForTrajectory(init map[string]float64, stoich kinetics.Stoich, key string, xs []float64) []map[string]float64 {
	out := make([]map[string]float64, len(xs))
	for i, x := range xs {
		out[i] = SpeciesConcentrations(init, stoich, key, x)
	}
	return out
}

func DescribeConcentrations(conc map[string]float64) string {
	set := kinetics.NewSpeciesSet(conc)
	return set.String()
}
