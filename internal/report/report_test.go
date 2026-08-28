package report_test

import (
	"encoding/json"
	"strings"
	"testing"

	"batch-arrh/internal/reactor"
	"batch-arrh/internal/report"
)

func buildResult(t *testing.T) reactor.Result {
	t.Helper()
	cfg := reactor.DefaultConfig()
	if err := json.Unmarshal([]byte(`{
	  "initial_concentration": {"A": 2.0},
	  "stoichiometry": {"A": -1, "P": 1},
	  "rate": {"order": "first", "k": 0.05},
	  "residence_time": 100.0
	}`), &cfg); err != nil {
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

func TestReportShowsConversionAndConcentrations(t *testing.T) {
	res := buildResult(t)
	out := report.Render(&res, report.Options{})
	for _, want := range []string{"conversion X_A = 0.993262", "C_A = 0.0134759", "C_P = 1.98652"} {
		if !strings.Contains(out, want) {
			t.Errorf("report missing %q; got:\n%s", want, out)
		}
	}
}

func TestReportTrajectoryStartsAtInitialState(t *testing.T) {
	res := buildResult(t)
	out := report.Render(&res, report.Options{Table: true, Rows: 2})
	firstRow := ""
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "0.0000") && strings.Contains(line, "0.00000") {
			firstRow = line
			break
		}
	}
	if firstRow == "" {
		t.Fatalf("trajectory table has no t=0 row; got:\n%s", out)
	}
	if !strings.Contains(firstRow, "2") {
		t.Errorf("t=0 row %q should show the initial concentration 2", firstRow)
	}
}

func TestJSONReportCarriesConversion(t *testing.T) {
	res := buildResult(t)
	data, err := report.RenderJSON(&res)
	if err != nil {
		t.Fatalf("RenderJSON returned error: %v", err)
	}
	var decoded struct {
		Conversion     float64            `json:"conversion"`
		Key            string             `json:"key_species"`
		Concentrations map[string]float64 `json:"concentrations"`
		Trajectory     []json.RawMessage  `json:"trajectory"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("JSON output does not parse: %v\n%s", err, data)
	}
	if decoded.Conversion != res.Conversion {
		t.Errorf("JSON conversion = %g, want %g", decoded.Conversion, res.Conversion)
	}
	if decoded.Key != "A" {
		t.Errorf("JSON key_species = %q, want A", decoded.Key)
	}
	if len(decoded.Trajectory) != len(res.Trajectory) {
		t.Errorf("JSON trajectory rows = %d, want %d", len(decoded.Trajectory), len(res.Trajectory))
	}
}

func TestJSONReportSeriesIncludesSelectivity(t *testing.T) {
	cfg := reactor.DefaultConfig()
	if err := json.Unmarshal([]byte(`{
	  "initial_concentration": {"A": 1.0},
	  "stoichiometry": {"A": -1, "B": 1, "C": 1},
	  "rate": {"order": "series", "k1": 0.10, "k2": 0.02},
	  "residence_time": 60.0
	}`), &cfg); err != nil {
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
	data, err := report.RenderJSON(&res)
	if err != nil {
		t.Fatalf("RenderJSON returned error: %v", err)
	}
	var decoded struct {
		Selectivity float64 `json:"selectivity"`
		SeriesPeakT float64 `json:"series_peak_time"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("JSON output does not parse: %v", err)
	}
	if decoded.Selectivity <= 0 || decoded.Selectivity > 1 {
		t.Errorf("JSON selectivity = %g, want inside (0,1]", decoded.Selectivity)
	}
	if decoded.SeriesPeakT <= 0 {
		t.Errorf("JSON series_peak_time = %g, want > 0", decoded.SeriesPeakT)
	}
}
