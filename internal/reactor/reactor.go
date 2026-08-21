package reactor

import (
	"errors"
	"fmt"

	"batch-arrh/internal/kinetics"
	"batch-arrh/internal/thermo"
)

// BatchReactor holds a validated scenario and the resolved kinetic
// parameters. It owns the integration state so the same configuration can
// be solved repeatedly without re-parsing the JSON.
type BatchReactor struct {
	cfg   Config
	order kinetics.Order
	k     float64
	key   string
	cA0   float64
	steps int
	vol   float64
	arr   *thermo.Arrhenius
}

// New validates a Config, resolves the rate constant (either direct or via
// the Arrhenius law) and prepares the solver state. Every illegal parameter
// reported in the spec surfaces here as an error.
func New(cfg *Config) (*BatchReactor, error) {
	if cfg == nil {
		return nil, errors.New("nil reactor config")
	}
	NormalizeConfig(cfg)
	if err := ValidateConfig(cfg); err != nil {
		return nil, err
	}
	order := kinetics.Order(cfg.Rate.Order)
	if !order.Valid() {
		return nil, fmt.Errorf("unknown rate order %q", cfg.Rate.Order)
	}
	r := &BatchReactor{
		cfg:   *cfg,
		order: order,
		key:   cfg.Key,
		cA0:   cfg.Initial[cfg.Key],
		steps: cfg.Steps,
		vol:   cfg.Volume,
	}
	if order == kinetics.OrderSeries {
		r.k = *cfg.Rate.K1
	} else {
		k, err := r.effectiveK()
		if err != nil {
			return nil, err
		}
		r.k = k
		r.arr = r.resolveArrhenius()
	}
	if err := kinetics.NewStoich(cfg.Stoich).Validate(); err != nil {
		return nil, err
	}
	if order == kinetics.OrderSecond {
		if _, err := r.secondReactant(kinetics.NewStoich(cfg.Stoich).Reactants()); err != nil {
			return nil, err
		}
	}
	return r, nil
}

// Solve runs the configured model to the batch residence time and returns
// the full trajectory plus the final summary.
func (r *BatchReactor) Solve() (Result, error) {
	if r.order == kinetics.OrderSeries {
		return r.solveSeries()
	}
	if r.order == kinetics.OrderFirst {
		return r.solveFirstOrder()
	}
	return r.solveSecondOrder()
}

// solveFirstOrder uses the closed form so the trajectory is exact on every
// grid point: X(t) = 1 - exp(-k*t).
func (r *BatchReactor) solveFirstOrder() (Result, error) {
	ts := timeGrid(0, r.cfg.Time, r.steps)
	xs := make([]float64, len(ts))
	for i, t := range ts {
		xs[i] = FirstOrderConversion(r.k, t)
	}
	return r.buildResult(ts, xs)
}

// solveSecondOrder integrates dX/dt with RK4. The derivative is built from
// the shared material balance, so the result honours the constant-volume
// relation and the rate law simultaneously.
func (r *BatchReactor) solveSecondOrder() (Result, error) {
	f, err := r.buildConversionDerivative()
	if err != nil {
		return Result{}, err
	}
	ts := timeGrid(0, r.cfg.Time, r.steps)
	xs, err := rk4(f, 0, 0, r.cfg.Time, r.steps)
	if err != nil {
		return Result{}, err
	}
	for i := range xs {
		xs[i] = clampConversion(xs[i])
	}
	return r.buildResult(ts, xs)
}

// solveSeries delegates to the consecutive-network solver and packages the
// snapshots into the common Result shape.
func (r *BatchReactor) solveSeries() (Result, error) {
	k2 := *r.cfg.Rate.K2
	pts, err := SolveSeries(r.cA0, r.k, k2, r.cfg.Time, r.steps)
	if err != nil {
		return Result{}, err
	}
	trajectory := make([]Point, len(pts))
	for i, p := range pts {
		trajectory[i] = Point{
			Time: p.Time,
			X:    ConversionFromConcentration(p.CA, r.cA0),
			C:    map[string]float64{"A": p.CA, "B": p.CB, "C": p.CC},
		}
	}
	last := pts[len(pts)-1]
	res := Result{
		Key:             r.key,
		Conversion:      ConversionFromConcentration(last.CA, r.cA0),
		Concentrations:  map[string]float64{"A": last.CA, "B": last.CB, "C": last.CC},
		Selectivity:     last.Selectivity,
		HasSelectivity:  true,
		SelectivityName: r.cfg.Selectivity,
		SeriesPeakTime:  SeriesPeakTime(r.k, k2),
		SeriesPeakConc:  SeriesPeakConcentration(r.cA0, r.k, k2),
		K:               r.k,
		K2:              k2,
		Order:           r.order,
		Arrhenius:       r.arr,
		Volume:          r.vol,
		Time:            r.cfg.Time,
		Steps:           r.steps,
		Trajectory:      trajectory,
	}
	return res, nil
}

// buildResult wraps a time grid and its conversion values into a Result
// with the final concentrations and (for series) the selectivity.
func (r *BatchReactor) buildResult(ts, xs []float64) (Result, error) {
	stoich := kinetics.NewStoich(r.cfg.Stoich)
	trajectory := make([]Point, len(xs))
	for i, x := range xs {
		trajectory[i] = Point{
			Time: ts[i],
			X:    x,
			C:    SpeciesConcentrations(r.cfg.Initial, stoich, r.key, x),
		}
	}
	last := trajectory[len(trajectory)-1]
	res := Result{
		Key:            r.key,
		Conversion:     last.X,
		Concentrations: last.C,
		K:              r.k,
		Order:          r.order,
		Arrhenius:      r.arr,
		Volume:         r.vol,
		Time:           r.cfg.Time,
		Steps:          r.steps,
		Trajectory:     trajectory,
	}
	if r.order == kinetics.OrderSecond {
		res.SecondOther = r.secondName()
	}
	return res, nil
}

// secondName resolves the co-reactant name for the report labels.
func (r *BatchReactor) secondName() string {
	stoich := kinetics.NewStoich(r.cfg.Stoich)
	for _, name := range stoich.Reactants() {
		if name != r.key {
			return name
		}
	}
	return ""
}
