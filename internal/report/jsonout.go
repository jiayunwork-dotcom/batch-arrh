package report

import (
	"encoding/json"
	"fmt"
	"sort"

	"batch-arrh/internal/reactor"
)

// jsonPoint is the machine-readable shape of one trajectory snapshot.
type jsonPoint struct {
	T float64            `json:"t"`
	X float64            `json:"x"`
	C map[string]float64 `json:"c"`
}

// jsonResult is the machine-readable summary of one integration. Field
// names are stable so external tooling can rely on the schema.
type jsonResult struct {
	Key            string             `json:"key_species"`
	Conversion     float64            `json:"conversion"`
	Concentrations map[string]float64 `json:"concentrations"`
	Selectivity    *float64           `json:"selectivity,omitempty"`
	SeriesPeakT    *float64           `json:"series_peak_time,omitempty"`
	SeriesPeakC    *float64           `json:"series_peak_concentration,omitempty"`
	Order          string             `json:"rate_order"`
	K              float64            `json:"rate_constant"`
	Volume         float64            `json:"volume"`
	Time           float64            `json:"residence_time"`
	Trajectory     []jsonPoint        `json:"trajectory"`
}

// RenderJSON marshals a result into the documented JSON schema. The
// trajectory is always included so a downstream consumer can re-plot the
// batch without re-running the solver.
func RenderJSON(res *reactor.Result) ([]byte, error) {
	jr := jsonResult{
		Key:            res.Key,
		Conversion:     res.Conversion,
		Concentrations: sortedConcentrations(res.Concentrations),
		Order:          string(res.Order),
		K:              res.K,
		Volume:         res.Volume,
		Time:           res.Time,
	}
	if res.HasSelectivity {
		v := res.Selectivity
		jr.Selectivity = &v
	}
	if res.IsSeries() {
		p := res.SeriesPeakTime
		q := res.SeriesPeakConc
		jr.SeriesPeakT = &p
		jr.SeriesPeakC = &q
	}
	jr.Trajectory = make([]jsonPoint, len(res.Trajectory))
	for i, p := range res.Trajectory {
		jr.Trajectory[i] = jsonPoint{T: p.Time, X: p.X, C: sortedConcentrations(p.C)}
	}
	return json.MarshalIndent(jr, "", "  ")
}

// sortedConcentrations returns a concentration map with a stable key order
// so repeated JSON marshalling produces byte-identical output.
func sortedConcentrations(c map[string]float64) map[string]float64 {
	out := make(map[string]float64, len(c))
	names := make([]string, 0, len(c))
	for name := range c {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		out[name] = c[name]
	}
	return out
}

// JSONError renders an error payload for machine-readable failure modes.
func JSONError(msg string) []byte {
	return []byte(fmt.Sprintf(`{"error": %q}`, msg))
}
