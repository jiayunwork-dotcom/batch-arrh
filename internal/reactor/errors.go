package reactor

import (
	"errors"
	"fmt"

	"batch-arrh/internal/kinetics"
)

var ErrNegativeTime = errors.New("residence time must be >= 0")

var ErrNonPositiveVolume = errors.New("volume must be > 0 L")

var ErrNonPositiveSteps = errors.New("time_steps must be >= 1")

var ErrMissingRate = errors.New("rate: supply k or arrhenius, not neither")

var ErrAmbiguousRate = errors.New("rate: k and arrhenius are mutually exclusive")

var ErrNegativeInitialConcentration = errors.New("initial concentration must be >= 0")

var ErrNonPositiveKeyConcentration = errors.New("key species must start with concentration > 0")

var ErrNonPositiveSeriesConstants = errors.New("series reaction needs k1 > 0 and k2 > 0")

func ValidateConfig(cfg *Config) error {
	if cfg.Time < 0 {
		return ErrNegativeTime
	}
	if cfg.Volume <= 0 {
		return ErrNonPositiveVolume
	}
	if cfg.Steps < 1 {
		return ErrNonPositiveSteps
	}
	if len(cfg.Initial) == 0 {
		return errors.New("initial_concentration must list at least one species")
	}
	for name, c := range cfg.Initial {
		if c < 0 {
			return fmt.Errorf("%w: %s = %g", ErrNegativeInitialConcentration, name, c)
		}
	}
	if cfg.Key == "" {
		cfg.Key = "A"
	}
	if cfg.Initial[cfg.Key] <= 0 {
		return fmt.Errorf("%w: %s = %g", ErrNonPositiveKeyConcentration, cfg.Key, cfg.Initial[cfg.Key])
	}
	return validateRate(&cfg.Rate)
}

func validateRate(r *RateSpec) error {
	switch r.Order {
	case "first", "second":
		hasK := r.K != nil
		hasArr := r.Arrhenius != nil
		if hasK && hasArr {
			return ErrAmbiguousRate
		}
		if hasArr {
			return thermoValidate(r.Arrhenius)
		}
		if !hasK {
			return ErrMissingRate
		}
		if *r.K < 0 {
			return fmt.Errorf("%w: k = %g", kinetics.ErrNegativeRateConstant, *r.K)
		}
		return nil
	case "series":
		if r.Arrhenius != nil {
			return ErrAmbiguousRate
		}
		if r.K1 == nil || r.K2 == nil {
			return ErrMissingRate
		}
		if *r.K1 <= 0 || *r.K2 <= 0 {
			return ErrNonPositiveSeriesConstants
		}
		return nil
	default:
		return fmt.Errorf("unknown rate order %q (want first, second or series)", r.Order)
	}
}
