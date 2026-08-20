package reactor

import (
	"batch-arrh/internal/thermo"
)

// thermoValidate bridges the reactor's validation into the Arrhenius
// parameter checks. It keeps the temperature and activation-energy rules in
// one place: T <= 0, Ea < 0 and A <= 0 are rejected identically whether the
// rate comes from a direct constant or from an Arrhenius triplet.
func thermoValidate(p *thermo.Arrhenius) error {
	return thermo.ValidateArrhenius(*p)
}

// NormalizeConfig fills the documented defaults into a freshly decoded
// Config. Explicit scenario values always win over defaults; omitted keys
// get the default so downstream code never has to branch on zero values.
func NormalizeConfig(cfg *Config) {
	if len(cfg.Stoich) == 0 {
		cfg.Stoich = map[string]float64{"A": -1, "P": 1}
	}
	if cfg.Volume <= 0 {
		cfg.Volume = 1.0
	}
	if cfg.Steps < 1 {
		cfg.Steps = 2000
	}
	if cfg.Key == "" {
		cfg.Key = "A"
	}
	if cfg.Selectivity == "" {
		cfg.Selectivity = "B"
	}
}

// effectiveK resolves the rate constant of a first or second order step.
// It prefers the Arrhenius evaluation when present and falls back to the
// direct value otherwise.
func (r *BatchReactor) effectiveK() (float64, error) {
	if r.cfg.Rate.Arrhenius != nil {
		return r.cfg.Rate.Arrhenius.Rate()
	}
	if r.cfg.Rate.K != nil {
		return *r.cfg.Rate.K, nil
	}
	return 0, ErrMissingRate
}

// resolveArrhenius returns the Arrhenius triplet attached to the rate, or
// nil when the scenario used a direct rate constant. The report uses this
// to disclose which parameters produced k.
func (r *BatchReactor) resolveArrhenius() *thermo.Arrhenius {
	return r.cfg.Rate.Arrhenius
}
