package reactor

import (
	"math"
	"sort"

	"batch-arrh/internal/kinetics"
	"batch-arrh/internal/thermo"
)

type Point struct {
	Time float64
	X    float64
	C    map[string]float64
}

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

func (res *Result) SpeciesOrder() []string {
	names := make([]string, 0, len(res.Concentrations))
	for name := range res.Concentrations {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (res *Result) ConversionRemainderAt() float64 {
	return 1 - res.Conversion
}

func (res *Result) RateConstant() float64 {
	return res.K
}

func (res *Result) SecondRateConstant() float64 {
	return res.K2
}

func (res *Result) IsSeries() bool {
	return res.Order == kinetics.OrderSeries
}

func (res *Result) InitialKeyConcentration() float64 {
	return res.Concentrations[res.Key] / (1 - res.Conversion)
}

func (res *Result) FinalConcentration(name string) float64 {
	return res.Concentrations[name]
}

func (res *Result) NKeyAtEnd() float64 {
	return res.Concentrations[res.Key] * res.Volume
}

func (res *Result) NKeyAtStart() float64 {
	return res.InitialKeyConcentration() * res.Volume
}

func (res *Result) SpeciesMoles() map[string]float64 {
	out := make(map[string]float64, len(res.Concentrations))
	for name, c := range res.Concentrations {
		out[name] = c * res.Volume
	}
	return out
}

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
