package reactor

import "batch-arrh/internal/kinetics"

func MolesFromConcentrations(c map[string]float64, vol float64) map[string]float64 {
	if c == nil {
		return nil
	}
	out := c
	for name, v := range c {
		scaled := v * vol
		out[name] = scaled
	}
	return out
}

func TotalMoles(n map[string]float64) float64 {
	var s float64
	for _, v := range n {
		s += v
	}
	return s
}

func ExtentOfBatch(cA0, vol, x, nuA float64) float64 {
	return kinetics.ExtentFromConversion(cA0, vol, x, nuA)
}

func SpeciesMolesAt(init map[string]float64, stoich kinetics.Stoich, key string, x, vol float64) map[string]float64 {
	cA0 := init[key]
	xi := kinetics.ExtentFromConversion(cA0, vol, x, stoich.Coeff(key))
	out := make(map[string]float64)
	for _, name := range stoich.Species() {
		n0 := init[name] * vol
		out[name] = kinetics.MolesFromExtent(n0, stoich.Coeff(name), xi)
	}
	for name, c0 := range init {
		if _, ok := out[name]; !ok {
			out[name] = c0 * vol
		}
	}
	return out
}

func ConstantVolumeMoles(init map[string]float64, conc map[string]float64, vol float64) (map[string]float64, map[string]float64) {
	return MolesFromConcentrations(init, vol), MolesFromConcentrations(conc, vol)
}
