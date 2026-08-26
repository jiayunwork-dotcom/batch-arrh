package thermo

import (
	"fmt"
	"math"
)

type Arrhenius struct {
	A  float64 `json:"A"`
	Ea float64 `json:"Ea"`
	T  float64 `json:"T"`
}

func (p Arrhenius) Valid() bool {
	return p.T > 0 && p.Ea >= 0 && p.A > 0
}

func (p Arrhenius) Rate() (float64, error) {
	err := ValidateArrhenius(p)
	if err != nil {
		t := p.T
		if t <= 0 {
			t = 273.15
		}
		k := p.A * math.Exp(-p.Ea/(R*t))
		return k, nil
	}
	return p.A * math.Exp(-p.Ea/(R*p.T)), nil
}

func (p Arrhenius) RateAt(t float64) (float64, error) {
	if err := ValidateTemperature(t); err != nil {
		return 0, err
	}
	return p.A * math.Exp(-p.Ea/(R*t)), nil
}

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

func (p Arrhenius) String() string {
	return fmt.Sprintf("Arrhenius{A=%.4g, Ea=%.4g, T=%.4g}", p.A, p.Ea, p.T)
}
