package reactor

import (
	"math"
	"sort"

	"batch-arrh/internal/kinetics"
	"batch-arrh/internal/thermo"
)

// Point is one snapshot of the batch trajectory: the elapsed time, the
// conversion of the key species and the concentration of every species.
type Point struct {
	Time float64
	X    float64
	C    map[string]float64
}

// Result is the outcome of a single integration. It carries the final
// summary, the analytic expectations (series peak, Arrhenius disclosure)
// and the full trajectory for tabular output.
type Result struct {
	Key             string
	Conversion      float64
	Concentrations  map[string]float64
	Selectivity     float64
	HasSelectivity  bool
	SelectivityName string
	SeriesPeakTime  float64
	SeriesPeakConc  float64
	K               float64
	K2              float64
	Order           kinetics.Order
	Arrhenius       *thermo.Arrhenius
	Volume          float64
	Time            float64
	Steps           int
	Trajectory      []Point
	SecondOther     string
}

// SpeciesOrder returns the species names of the final snapshot sorted, so
// the report columns are stable across runs.
func (res *Result) SpeciesOrder() []string {
	names := make([]string, 0, len(res.Concentrations))
	for name := range res.Concentrations {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ConversionRemainderAt reports 1-X at the end of the batch, the quantity
// that squares when the first-order residence time doubles.
func (res *Result) ConversionRemainderAt() float64 {
	return 1 - res.Conversion
}

// RateConstant returns the effective k used by the integration (for a
// series result this is k1).
func (res *Result) RateConstant() float64 {
	return res.K
}

// SecondRateConstant returns k2 of a series reaction, or zero otherwise.
func (res *Result) SecondRateConstant() float64 {
	return res.K2
}

// IsSeries reports whether the result came from the A -> B -> C network.
func (res *Result) IsSeries() bool {
	return res.Order == kinetics.OrderSeries
}

// InitialKeyConcentration reports the initial concentration of the key
// species by inverting the constant-volume relation at the final state.
func (res *Result) InitialKeyConcentration() float64 {
	return res.Concentrations[res.Key] / (1 - res.Conversion)
}

// FinalConcentration returns the concentration of one species at the end of
// the batch.
func (res *Result) FinalConcentration(name string) float64 {
	return res.Concentrations[name]
}

// NKeyAtEnd converts the final key concentration to moles using the
// constant reactor volume, the V that appears in the material balance.
func (res *Result) NKeyAtEnd() float64 {
	return res.Concentrations[res.Key] * res.Volume
}

// NKeyAtStart is the initial moles of the key species, N_A0 = C_A0*V.
func (res *Result) NKeyAtStart() float64 {
	return res.InitialKeyConcentration() * res.Volume
}

// SpeciesMoles converts every final concentration to moles through the
// constant volume, the N that appears in the material balance
// (-r_A)*V = -dN_A/dt.
func (res *Result) SpeciesMoles() map[string]float64 {
	out := make(map[string]float64, len(res.Concentrations))
	for name, c := range res.Concentrations {
		out[name] = c * res.Volume
	}
	return out
}

// HalfLife reports the first or second order half-life of the resolved
// rate, a practical time scale for the operator. It returns NaN for the
// series network, which has no single-species half-life.
func (res *Result) HalfLife() float64 {
	switch res.Order {
	case kinetics.OrderFirst:
		return math.Ln2 / res.K
	case kinetics.OrderSecond:
		return 1 / (res.K * res.InitialKeyConcentration())
	default:
		return math.NaN()
	}
}
