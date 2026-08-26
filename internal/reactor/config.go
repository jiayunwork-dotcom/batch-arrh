package reactor

import (
	"batch-arrh/internal/thermo"
)

type Config struct {
	Initial map[string]float64 `json:"initial_concentration"`

	Stoich map[string]float64 `json:"stoichiometry"`

	Rate RateSpec `json:"rate"`

	Volume float64 `json:"volume"`

	Time float64 `json:"residence_time"`

	Steps int `json:"time_steps"`

	Key string `json:"key_species"`

	Selectivity string `json:"selectivity_species"`
}

type RateSpec struct {
	Order string `json:"order"`

	K *float64 `json:"k"`

	K1 *float64 `json:"k1"`
	K2 *float64 `json:"k2"`

	Arrhenius *thermo.Arrhenius `json:"arrhenius"`
}

func DefaultConfig() Config {
	return Config{
		Stoich:      map[string]float64{"A": -1, "P": 1},
		Volume:      1.0,
		Steps:       2000,
		Key:         "A",
		Selectivity: "B",
	}
}
