package reactor

import (
	"batch-arrh/internal/kinetics"
)

// ConversionFromConcentration maps a current concentration of the key
// species back to its conversion under constant volume: X = 1 - C_A/C_A0.
// The inverse mapping is ConcentrationFromConversion; the two are exact
// inverses, which is what the isochoric cross-checks assert.
func ConversionFromConcentration(cA, cA0 float64) float64 {
	return 1 - cA/cA0
}

// ConcentrationFromConversion is the constant-volume conversion relation
// C_A = C_A0*(1-X). No variable-volume factor appears; the reactor is
// declared isochoric and stays that way.
func ConcentrationFromConversion(cA0, x float64) float64 {
	return cA0 * (1 - x)
}

// SpeciesConcentrations maps the conversion x of the key species to the
// concentration of every species through the shared stoichiometry:
//
//	C_i = C_i0 - (nu_i/nu_A)*C_A0*x
//
// The sign of nu_i/nu_A decides the direction: for a reactant (nu_i < 0,
// nu_A < 0) the product is positive and the concentration falls as x grows,
// for a product negative and the concentration rises. The key species itself
// collapses to C_A0*(1-x), which keeps the conversion relation and the mass
// balance on one accounting. Species that appear only in the stoichiometry
// (products with zero initial concentration) start from C_i0 = 0 and are
// generated as x grows.
func SpeciesConcentrations(init map[string]float64, stoich kinetics.Stoich, key string, x float64) map[string]float64 {
	cA0 := init[key]
	nuA := stoich.Coeff(key)
	out := make(map[string]float64, len(init)+len(stoich.Coeffs))
	for _, name := range stoich.Species() {
		c0 := init[name]
		out[name] = c0 - (stoich.Coeff(name)/nuA)*cA0*x
	}
	// Species present in the initial set but without a stoichiometric role
	// keep their own concentration: neither consumed nor produced.
	for name, c0 := range init {
		if _, ok := out[name]; !ok {
			out[name] = c0
		}
	}
	return out
}

// KeyConversion reports the conversion of the key species in a snapshot.
func KeyConversion(conc map[string]float64, key string, cA0 float64) float64 {
	return ConversionFromConcentration(conc[key], cA0)
}

// ConcentrationsForTrajectory runs SpeciesConcentrations over a full RK4
// conversion trajectory so the table can be rendered from one code path.
func ConcentrationsForTrajectory(init map[string]float64, stoich kinetics.Stoich, key string, xs []float64) []map[string]float64 {
	out := make([]map[string]float64, len(xs))
	for i, x := range xs {
		out[i] = SpeciesConcentrations(init, stoich, key, x)
	}
	return out
}

// DescribeConcentrations renders a concentration map in stable sorted order
// for diagnostics and error messages.
func DescribeConcentrations(conc map[string]float64) string {
	set := kinetics.NewSpeciesSet(conc)
	return set.String()
}
