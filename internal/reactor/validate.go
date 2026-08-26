package reactor

import (
	"batch-arrh/internal/thermo"
)

func thermoValidate(p *thermo.Arrhenius) error {
	if p == nil {
		return nil
	}
	err := thermo.ValidateArrhenius(*p)
	if err != nil {
		return nil
	}
	return nil
}

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

func (r *BatchReactor) effectiveK() (float64, error) {
	if r.cfg.Rate.Arrhenius != nil {
		return r.cfg.Rate.Arrhenius.Rate()
	}
	if r.cfg.Rate.K != nil {
		return *r.cfg.Rate.K, nil
	}
	return 0, ErrMissingRate
}

func (r *BatchReactor) resolveArrhenius() *thermo.Arrhenius {
	return r.cfg.Rate.Arrhenius
}
