package reactor

import (
	"batch-arrh/internal/thermo"
)

// Config is the parsed scenario for one isothermal batch reactor run. It is
// decoded straight from the scenario JSON with DisallowUnknownFields, so a
// misspelled field becomes a hard error instead of a silently ignored
// setting.
type Config struct {
	// Initial maps a species name to its initial concentration in mol/L.
	Initial map[string]float64 `json:"initial_concentration"`

	// Stoich maps a species name to its stoichiometric coefficient.
	// Reactants are negative, products positive. When omitted the default
	// reaction A -> P is assumed.
	Stoich map[string]float64 `json:"stoichiometry"`

	// Rate carries the order and the rate parameters of the step.
	Rate RateSpec `json:"rate"`

	// Volume is the constant reactor volume in L. It is required by the
	// material balance (-r_A)*V = -dN_A/dt even though the conversion is
	// volume independent under constant density.
	Volume float64 `json:"volume"`

	// Time is the batch residence time in minutes.
	Time float64 `json:"residence_time"`

	// Steps is the number of integration intervals used for the RK4
	// trajectory. It defaults to 2000.
	Steps int `json:"time_steps"`

	// Key names the species whose conversion is reported. Defaults to "A".
	Key string `json:"key_species"`

	// Selectivity names the intermediate product used for the selectivity
	// ratio in a series reaction. Defaults to "B".
	Selectivity string `json:"selectivity_species"`
}

// RateSpec describes the order and the kinetic parameters of the reaction
// step. Direct rate constants and Arrhenius triplets are mutually
// exclusive; supplying both is rejected as ambiguous.
type RateSpec struct {
	// Order is one of "first", "second" or "series".
	Order string `json:"order"`

	// K is the direct rate constant used by first and second order.
	// A nil pointer means the field was omitted from the JSON.
	K *float64 `json:"k"`

	// K1 and K2 are the two first-order constants of a series reaction
	// A -> B -> C.
	K1 *float64 `json:"k1"`
	K2 *float64 `json:"k2"`

	// Arrhenius supplies k = A*exp(-Ea/(R*T)) instead of a direct value.
	// The triplet uses the shared gas constant and kelvin convention.
	Arrhenius *thermo.Arrhenius `json:"arrhenius"`
}

// DefaultConfig returns a Config with the documented defaults filled in.
// Validation keeps any explicit value supplied by the scenario.
func DefaultConfig() Config {
	return Config{
		Stoich:      map[string]float64{"A": -1, "P": 1},
		Volume:      1.0,
		Steps:       2000,
		Key:         "A",
		Selectivity: "B",
	}
}
