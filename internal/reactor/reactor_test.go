package reactor_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"testing"

	"batch-arrh/internal/kinetics"
	"batch-arrh/internal/reactor"
	"batch-arrh/internal/thermo"
)

func solveScenario(t *testing.T, scenario string) reactor.Result {
	t.Helper()
	cfg := reactor.DefaultConfig()
	if err := json.Unmarshal([]byte(scenario), &cfg); err != nil {
		t.Fatalf("scenario does not parse: %v", err)
	}
	r, err := reactor.New(&cfg)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	res, err := r.Solve()
	if err != nil {
		t.Fatalf("Solve returned error: %v", err)
	}
	return res
}

func newErr(t *testing.T, scenario string) error {
	t.Helper()
	cfg := reactor.DefaultConfig()
	if err := json.Unmarshal([]byte(scenario), &cfg); err != nil {
		t.Fatalf("scenario does not parse: %v", err)
	}
	_, err := reactor.New(&cfg)
	return err
}

const firstOrderJSON = `{
  "initial_concentration": {"A": 2.0},
  "stoichiometry": {"A": -1, "P": 1},
  "rate": {"order": "first", "k": 0.05},
  "residence_time": 100.0
}`

func TestFirstOrderMatchesClosedForm(t *testing.T) {
	res := solveScenario(t, firstOrderJSON)
	wantX := 1 - math.Exp(-0.05*100)
	if math.Abs(res.Conversion-wantX) > 1e-12 {
		t.Errorf("X_A = %.12g, want %.12g", res.Conversion, wantX)
	}
	wantCA := 2.0 * math.Exp(-0.05*100)
	if math.Abs(res.Concentrations["A"]-wantCA) > 1e-12 {
		t.Errorf("C_A = %.12g, want %.12g", res.Concentrations["A"], wantCA)
	}
}

func TestZeroTimeReturnsInitial(t *testing.T) {
	scenario := `{
	  "initial_concentration": {"A": 2.0},
	  "stoichiometry": {"A": -1, "P": 1},
	  "rate": {"order": "first", "k": 0.05},
	  "residence_time": 0.0
	}`
	res := solveScenario(t, scenario)
	if res.Conversion != 0 {
		t.Errorf("X_A = %g, want 0 at t = 0", res.Conversion)
	}
	if res.Concentrations["A"] != 2.0 {
		t.Errorf("C_A = %g, want 2.0 at t = 0", res.Concentrations["A"])
	}
	if res.Concentrations["P"] != 0 {
		t.Errorf("C_P = %g, want 0 at t = 0", res.Concentrations["P"])
	}
	if res.Trajectory[0].Time != 0 {
		t.Errorf("first trajectory point at t = %g, want 0", res.Trajectory[0].Time)
	}
}

func TestDoublingTimeSquaresRemainder(t *testing.T) {
	scenario := `{
	  "initial_concentration": {"A": 1.0},
	  "stoichiometry": {"A": -1, "P": 1},
	  "rate": {"order": "first", "k": 0.1},
	  "residence_time": 10.0
	}`
	cfg := reactor.DefaultConfig()
	if err := json.Unmarshal([]byte(scenario), &cfg); err != nil {
		t.Fatalf("scenario does not parse: %v", err)
	}
	r, err := reactor.New(&cfg)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	short, err := r.Solve()
	if err != nil {
		t.Fatalf("Solve at t=10 returned error: %v", err)
	}
	cfg.Time = 20.0
	r2, err := reactor.New(&cfg)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	long, err := r2.Solve()
	if err != nil {
		t.Fatalf("Solve at t=20 returned error: %v", err)
	}
	got := 1 - long.Conversion
	want := (1 - short.Conversion) * (1 - short.Conversion)
	if math.Abs(got-want) > 1e-12 {
		t.Errorf("1-X(20) = %.12g, want (1-X(10))^2 = %.12g", got, want)
	}
}

func TestRisingTemperatureRaisesConversion(t *testing.T) {
	base := `{
	  "initial_concentration": {"A": 2.0},
	  "stoichiometry": {"A": -1, "P": 1},
	  "rate": {"order": "first", "arrhenius": {"A": 100.0, "Ea": 20000.0, "T": %g}},
	  "residence_time": 20.0
	}`
	cold := solveScenario(t, fmt.Sprintf(base, 300.0))
	hot := solveScenario(t, fmt.Sprintf(base, 320.0))
	if hot.Conversion <= cold.Conversion {
		t.Errorf("X(320 K) = %g, want > X(300 K) = %g", hot.Conversion, cold.Conversion)
	}
	if hot.K <= cold.K {
		t.Errorf("k(320 K) = %g, want > k(300 K) = %g", hot.K, cold.K)
	}
}

const secondOrderJSON = `{
  "initial_concentration": {"A": 2.0, "B": 2.0},
  "stoichiometry": {"A": -1, "B": -1, "P": 1},
  "rate": {"order": "second", "k": 0.01},
  "residence_time": 100.0
}`

func TestSecondOrderMatchesInverseConcentration(t *testing.T) {
	res := solveScenario(t, secondOrderJSON)
	got := 1/res.Concentrations["A"] - 1/2.0
	want := 0.01 * 100.0
	if math.Abs(got-want) > 1e-8 {
		t.Errorf("1/C_A - 1/C_A0 = %.10g, want k*t = %g", got, want)
	}
}

func TestSecondOrderConstantVolumeRelation(t *testing.T) {
	res := solveScenario(t, secondOrderJSON)
	last := res.Trajectory[len(res.Trajectory)-1]
	want := 2.0 * (1 - last.X)
	if math.Abs(last.C["A"]-want) > 1e-9 {
		t.Errorf("C_A(X) = %g, want C_A0*(1-X) = %g", last.C["A"], want)
	}
	for i, p := range res.Trajectory {
		expect := 2.0 * (1 - p.X)
		if math.Abs(p.C["A"]-expect) > 1e-9 {
			t.Errorf("trajectory[%d]: C_A = %g, want %g", i, p.C["A"], expect)
		}
	}
}

func TestSecondOrderConversionClosedForm(t *testing.T) {
	res := solveScenario(t, secondOrderJSON)
	want := 0.01 * 2.0 * 100.0 / (1 + 0.01*2.0*100.0)
	if math.Abs(res.Conversion-want) > 1e-8 {
		t.Errorf("X_A = %.10g, want %.10g", res.Conversion, want)
	}
}

const seriesJSON = `{
  "initial_concentration": {"A": 1.0},
  "stoichiometry": {"A": -1, "B": 1, "C": 1},
  "rate": {"order": "series", "k1": 0.10, "k2": 0.02},
  "residence_time": 60.0
}`

func TestSeriesConservesMoles(t *testing.T) {
	res := solveScenario(t, seriesJSON)
	if len(res.Trajectory) == 0 {
		t.Fatal("series result has an empty trajectory")
	}
	for i, p := range res.Trajectory {
		sum := p.C["A"] + p.C["B"] + p.C["C"]
		if math.Abs(sum-1.0) > 1e-9 {
			t.Errorf("trajectory[%d]: A+B+C = %.10g, want C_A0 = 1", i, sum)
		}
	}
}

func TestSeriesPeakTimeMatchesAnalytic(t *testing.T) {
	res := solveScenario(t, seriesJSON)
	want := math.Log(0.10/0.02) / (0.10 - 0.02)
	best := res.Trajectory[0]
	for _, p := range res.Trajectory {
		if p.C["B"] > best.C["B"] {
			best = p
		}
	}
	step := 60.0 / 2000.0
	if math.Abs(best.Time-want) > 2*step {
		t.Errorf("trajectory peak at t = %g, want analytic t_max = %g", best.Time, want)
	}
	if math.Abs(res.SeriesPeakTime-want) > 1e-9 {
		t.Errorf("SeriesPeakTime = %g, want %g", res.SeriesPeakTime, want)
	}
}

func TestSeriesPeakConcentrationMatchesAnalytic(t *testing.T) {
	res := solveScenario(t, seriesJSON)
	maxCB := 0.0
	for _, p := range res.Trajectory {
		if p.C["B"] > maxCB {
			maxCB = p.C["B"]
		}
	}
	want := res.SeriesPeakConc
	if math.Abs(maxCB-want) > 1e-6 {
		t.Errorf("max C_B = %.9g, want analytic %.9g", maxCB, want)
	}
}

func TestSeriesSelectivityInUnitInterval(t *testing.T) {
	res := solveScenario(t, seriesJSON)
	if !res.HasSelectivity {
		t.Fatal("series result should carry a selectivity")
	}
	if res.Selectivity < 0 || res.Selectivity > 1 {
		t.Errorf("final selectivity = %g, want inside [0,1]", res.Selectivity)
	}
}

func TestNegativeKRejected(t *testing.T) {
	scenario := `{
	  "initial_concentration": {"A": 2.0},
	  "stoichiometry": {"A": -1, "P": 1},
	  "rate": {"order": "first", "k": -0.05},
	  "residence_time": 100.0
	}`
	err := newErr(t, scenario)
	if err == nil {
		t.Fatal("New accepted k < 0, want error")
	}
	if !errors.Is(err, kinetics.ErrNegativeRateConstant) {
		t.Errorf("err = %v, want the negative rate constant error", err)
	}
}

func TestZeroTemperatureRejected(t *testing.T) {
	scenario := `{
	  "initial_concentration": {"A": 2.0},
	  "stoichiometry": {"A": -1, "P": 1},
	  "rate": {"order": "first", "arrhenius": {"A": 1e12, "Ea": 80000.0, "T": 0}},
	  "residence_time": 100.0
	}`
	err := newErr(t, scenario)
	if err == nil {
		t.Fatal("New accepted T = 0, want error")
	}
	if !errors.Is(err, thermo.ErrNonPositiveTemperature) {
		t.Errorf("err = %v, want ErrNonPositiveTemperature", err)
	}
}

func TestNegativeInitialConcentrationRejected(t *testing.T) {
	scenario := `{
	  "initial_concentration": {"A": -2.0},
	  "stoichiometry": {"A": -1, "P": 1},
	  "rate": {"order": "first", "k": 0.05},
	  "residence_time": 100.0
	}`
	err := newErr(t, scenario)
	if err == nil {
		t.Fatal("New accepted a negative initial concentration, want error")
	}
}

func TestNegativeTimeRejected(t *testing.T) {
	scenario := `{
	  "initial_concentration": {"A": 2.0},
	  "stoichiometry": {"A": -1, "P": 1},
	  "rate": {"order": "first", "k": 0.05},
	  "residence_time": -5.0
	}`
	err := newErr(t, scenario)
	if err == nil {
		t.Fatal("New accepted t < 0, want error")
	}
	if !errors.Is(err, reactor.ErrNegativeTime) {
		t.Errorf("err = %v, want ErrNegativeTime", err)
	}
}

func TestAmbiguousRateRejected(t *testing.T) {
	scenario := `{
	  "initial_concentration": {"A": 2.0},
	  "stoichiometry": {"A": -1, "P": 1},
	  "rate": {"order": "first", "k": 0.05, "arrhenius": {"A": 1e12, "Ea": 80000.0, "T": 350}},
	  "residence_time": 100.0
	}`
	err := newErr(t, scenario)
	if err == nil {
		t.Fatal("New accepted both k and arrhenius, want error")
	}
	if !errors.Is(err, reactor.ErrAmbiguousRate) {
		t.Errorf("err = %v, want ErrAmbiguousRate", err)
	}
}

func TestMissingRateRejected(t *testing.T) {
	scenario := `{
	  "initial_concentration": {"A": 2.0},
	  "stoichiometry": {"A": -1, "P": 1},
	  "rate": {"order": "first"},
	  "residence_time": 100.0
	}`
	err := newErr(t, scenario)
	if err == nil {
		t.Fatal("New accepted a scenario without rate parameters, want error")
	}
	if !errors.Is(err, reactor.ErrMissingRate) {
		t.Errorf("err = %v, want ErrMissingRate", err)
	}
}

func TestSecondOrderNeedsTwoReactants(t *testing.T) {
	scenario := `{
	  "initial_concentration": {"A": 2.0},
	  "stoichiometry": {"A": -1, "P": 1},
	  "rate": {"order": "second", "k": 0.01},
	  "residence_time": 100.0
	}`
	err := newErr(t, scenario)
	if err == nil {
		t.Fatal("New accepted second-order with one reactant, want error")
	}
}

func TestArrheniusDisclosedInResult(t *testing.T) {
	scenario := `{
	  "initial_concentration": {"A": 2.0},
	  "stoichiometry": {"A": -1, "P": 1},
	  "rate": {"order": "first", "arrhenius": {"A": 1e12, "Ea": 80000.0, "T": 350}},
	  "residence_time": 100.0
	}`
	res := solveScenario(t, scenario)
	if res.Arrhenius == nil {
		t.Fatal("result should disclose the Arrhenius triplet")
	}
	if res.Arrhenius.T != 350.0 {
		t.Errorf("disclosed T = %g, want 350", res.Arrhenius.T)
	}
}

func TestDamkohlerFirstOrderAgreesWithClosedForm(t *testing.T) {
	res := solveScenario(t, firstOrderJSON)
	da := reactor.FirstOrderDamkohler(0.05, 100)
	if da != 5 {
		t.Errorf("Da = %g, want 5", da)
	}
	want := reactor.ConversionFromFirstOrderDa(da)
	if math.Abs(res.Conversion-want) > 1e-12 {
		t.Errorf("X = %g, want Da closed form %g", res.Conversion, want)
	}
	tNeed, err := reactor.FirstOrderTimeForConversion(0.05, res.Conversion)
	if err != nil {
		t.Fatalf("FirstOrderTimeForConversion: %v", err)
	}
	if math.Abs(tNeed-100) > 1e-9 {
		t.Errorf("design t = %g, want 100", tNeed)
	}
}

func TestUnequalSecondOrderMatchesRK4(t *testing.T) {
	scenario := `{
	  "initial_concentration": {"A": 2.0, "B": 1.0},
	  "stoichiometry": {"A": -1, "B": -1, "P": 1},
	  "rate": {"order": "second", "k": 0.01},
	  "residence_time": 50.0,
	  "time_steps": 4000
	}`
	res := solveScenario(t, scenario)
	want := reactor.SecondOrderUnequalConversion(0.01, 2.0, 1.0, 50.0)
	if math.Abs(res.Conversion-want) > 1e-6 {
		t.Errorf("X_A = %.10g, want unequal closed form %.10g", res.Conversion, want)
	}
	wantCA := reactor.SecondOrderUnequalConcentrationA(2.0, 0.01, 1.0, 50.0)
	if math.Abs(res.Concentrations["A"]-wantCA) > 1e-6 {
		t.Errorf("C_A = %g, want %g", res.Concentrations["A"], wantCA)
	}
	wantCB := reactor.SecondOrderUnequalConcentrationB(2.0, 0.01, 1.0, 50.0)
	if math.Abs(res.Concentrations["B"]-wantCB) > 1e-6 {
		t.Errorf("C_B = %g, want %g", res.Concentrations["B"], wantCB)
	}
	xmax := reactor.LimitingConversion(2.0, 1.0, -1, -1)
	if xmax != 0.5 {
		t.Errorf("limiting X = %g, want 0.5", xmax)
	}
	if res.Conversion >= xmax {
		t.Errorf("X = %g reached or passed the B-limited ceiling %g", res.Conversion, xmax)
	}
}

func TestSeriesYieldEqualsSelectivityTimesConversion(t *testing.T) {
	res := solveScenario(t, seriesJSON)
	cA0 := 1.0
	for i, p := range res.Trajectory {
		y := reactor.IntermediateYield(p.C["B"], cA0)
		s := reactor.OverallSelectivity(p.C["B"], p.C["A"], cA0)
		got := reactor.YieldFromSelectivity(s, p.X)
		if math.Abs(y-got) > 1e-9 {
			t.Errorf("trajectory[%d]: Y_B = %g, S_B*X = %g", i, y, got)
		}
		sum := p.C["A"] + p.C["B"] + p.C["C"]
		if math.Abs(sum-cA0) > 1e-9 {
			t.Errorf("trajectory[%d]: A+B+C = %g, want 1", i, sum)
		}
	}
	peakY := reactor.SeriesPeakYield(cA0, 0.10, 0.02)
	if math.Abs(peakY-res.SeriesPeakConc/cA0) > 1e-12 {
		t.Errorf("peak yield %g, want %g", peakY, res.SeriesPeakConc/cA0)
	}
}

func TestMolesScaleWithVolumeAndExtent(t *testing.T) {
	res := solveScenario(t, firstOrderJSON)
	n := reactor.MolesFromConcentrations(res.Concentrations, 2.0)
	if math.Abs(n["A"]-res.Concentrations["A"]*2) > 1e-12 {
		t.Errorf("N_A = %g, want C_A*2", n["A"])
	}
	xi := reactor.ExtentOfBatch(2.0, 2.0, res.Conversion, -1)
	nA := kinetics.MolesFromExtent(4.0, -1, xi)
	if math.Abs(nA-n["A"]) > 1e-9 {
		t.Errorf("extent moles %g, want %g", nA, n["A"])
	}
}
