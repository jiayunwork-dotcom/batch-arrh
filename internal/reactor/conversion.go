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
	cA0 := init[key]
	nuA := stoich.Coeff(key)
	out := make(map[string]float64, len(init)+len(stoich.Coeffs))
	for _, name := range stoich.Species() {
		c0 := init[name]
		out[name] = c0 - (stoich.Coeff(name)/nuA)*cA0*x
	}
	for name, c0 := range init {
		if _, ok := out[name]; !ok {
			out[name] = c0
		}
	}
	return out
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
