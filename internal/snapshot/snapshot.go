package snapshot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"batch-arrh/internal/reactor"
	"batch-arrh/internal/thermo"
)

const (
	Magic          = "BAR1"
	CurrentVersion = 1
	Tol            = 1e-9
)

type Record struct {
	Magic          string             `json:"magic"`
	Version        int                `json:"version"`
	Order          string             `json:"order"`
	K              float64            `json:"k"`
	K2             float64            `json:"k2,omitempty"`
	Time           float64            `json:"residence_time"`
	Volume         float64            `json:"volume"`
	Steps          int                `json:"time_steps"`
	Key            string             `json:"key_species"`
	SelectivitySp  string             `json:"selectivity_species,omitempty"`
	Initial        map[string]float64 `json:"initial_concentration"`
	Stoich         map[string]float64 `json:"stoichiometry"`
	Conversion     float64            `json:"conversion"`
	Concentrations map[string]float64 `json:"concentrations"`
	HasArrhenius   bool               `json:"has_arrhenius"`
	ArrA           float64            `json:"arrhenius_a,omitempty"`
	ArrEa          float64            `json:"arrhenius_ea,omitempty"`
	ArrT           float64            `json:"arrhenius_t,omitempty"`
	Selectivity    float64            `json:"selectivity,omitempty"`
	SeriesPeakTime float64            `json:"series_peak_time,omitempty"`
}

func Capture(cfg reactor.Config) (Record, error) {
	cp := cloneConfig(cfg)
	rx, err := reactor.New(&cp)
	if err != nil {
		return Record{}, fmt.Errorf("snapshot: %w", err)
	}
	res, err := rx.Solve()
	if err != nil {
		return Record{}, fmt.Errorf("snapshot: %w", err)
	}
	return fromKernel(cp, res), nil
}

func fromKernel(cfg reactor.Config, res reactor.Result) Record {
	rec := Record{
		Magic:          Magic,
		Version:        CurrentVersion,
		Order:          string(res.Order),
		K:              res.K,
		K2:             res.K2,
		Time:           cfg.Time,
		Volume:         res.Volume,
		Steps:          res.Steps,
		Key:            res.Key,
		SelectivitySp:  cfg.Selectivity,
		Initial:        copyMap(cfg.Initial),
		Stoich:         copyMap(cfg.Stoich),
		Conversion:     res.Conversion,
		Concentrations: copyMap(res.Concentrations),
	}
	if cfg.Rate.Arrhenius != nil {
		rec.HasArrhenius = true
		rec.ArrA = cfg.Rate.Arrhenius.A
		rec.ArrEa = cfg.Rate.Arrhenius.Ea
		rec.ArrT = cfg.Rate.Arrhenius.T
	}
	if res.HasSelectivity {
		rec.Selectivity = res.Selectivity
	}
	if res.IsSeries() {
		rec.SeriesPeakTime = res.SeriesPeakTime
	}
	return rec
}

func (r Record) config() reactor.Config {
	cfg := reactor.DefaultConfig()
	cfg.Initial = copyMap(r.Initial)
	cfg.Stoich = copyMap(r.Stoich)
	cfg.Volume = r.Volume
	cfg.Time = r.Time
	if r.Steps > 0 {
		cfg.Steps = r.Steps
	}
	if r.Key != "" {
		cfg.Key = r.Key
	}
	if r.SelectivitySp != "" {
		cfg.Selectivity = r.SelectivitySp
	}
	cfg.Rate.Order = r.Order
	switch r.Order {
	case "series":
		k1 := r.K
		k2 := r.K2
		cfg.Rate.K1 = &k1
		cfg.Rate.K2 = &k2
	default:
		if r.HasArrhenius {
			cfg.Rate.Arrhenius = &thermo.Arrhenius{A: r.ArrA, Ea: r.ArrEa, T: r.ArrT}
		} else {
			k := r.K
			cfg.Rate.K = &k
		}
	}
	return cfg
}

func (r Record) validate() error {
	if r.Magic != Magic {
		return fmt.Errorf("snapshot: bad magic %q", r.Magic)
	}
	if r.Version != CurrentVersion {
		return fmt.Errorf("snapshot: unsupported version %d", r.Version)
	}
	if err := r.validateInputs(); err != nil {
		return err
	}
	if r.Conversion < 0 || r.Conversion > 1 {
		return fmt.Errorf("snapshot: stored conversion %g outside [0, 1]", r.Conversion)
	}
	if r.K < 0 {
		return fmt.Errorf("snapshot: stored k must be >= 0")
	}
	if r.Order == "series" && r.K2 <= 0 {
		return fmt.Errorf("snapshot: series record needs k2 > 0")
	}
	if len(r.Concentrations) == 0 {
		return fmt.Errorf("snapshot: stored concentrations missing")
	}
	return nil
}

func (r Record) validateInputs() error {
	cfg := r.config()
	if _, err := reactor.New(&cfg); err != nil {
		return fmt.Errorf("snapshot: %w", err)
	}
	return nil
}

func WriteFile(path string, rec Record) error {
	if rec.Magic == "" {
		rec.Magic = Magic
	}
	if rec.Version == 0 {
		rec.Version = CurrentVersion
	}
	if err := rec.validate(); err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	raw, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return err
	}
	if len(raw) == 0 {
		return fmt.Errorf("snapshot: empty marshal")
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func ReadFile(path string) (Record, error) {
	var rec Record
	raw, err := os.ReadFile(path)
	if err != nil {
		return rec, err
	}
	if len(raw) == 0 {
		return rec, fmt.Errorf("snapshot: empty file")
	}
	if !json.Valid(raw) {
		return rec, fmt.Errorf("snapshot: truncated or invalid JSON")
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&rec); err != nil {
		return Record{}, fmt.Errorf("snapshot: %w", err)
	}
	if dec.More() {
		return Record{}, fmt.Errorf("snapshot: trailing content")
	}
	if err := rec.validate(); err != nil {
		return Record{}, err
	}
	return rec, nil
}

func (r Record) ReplayAgrees() error {
	if err := r.validate(); err != nil {
		return err
	}
	live, err := Capture(r.config())
	if err != nil {
		return err
	}
	if math.Abs(live.Conversion-r.Conversion) > Tol {
		return fmt.Errorf("snapshot: live conversion %v stored %v", live.Conversion, r.Conversion)
	}
	if math.Abs(live.K-r.K) > Tol*math.Max(1, math.Abs(r.K)) {
		return fmt.Errorf("snapshot: live k %v stored %v", live.K, r.K)
	}
	if r.Order == "series" {
		if math.Abs(live.K2-r.K2) > Tol*math.Max(1, math.Abs(r.K2)) {
			return fmt.Errorf("snapshot: live k2 %v stored %v", live.K2, r.K2)
		}
		if math.Abs(live.SeriesPeakTime-r.SeriesPeakTime) > Tol*math.Max(1, math.Abs(r.SeriesPeakTime)) {
			return fmt.Errorf("snapshot: live peak time %v stored %v", live.SeriesPeakTime, r.SeriesPeakTime)
		}
	}
	for name, c := range r.Concentrations {
		got := live.Concentrations[name]
		scale := math.Max(1, math.Max(math.Abs(c), math.Abs(got)))
		if math.Abs(got-c) > Tol*scale {
			return fmt.Errorf("snapshot: live C_%s %v stored %v", name, got, c)
		}
	}
	return nil
}

func (r Record) Matches(other Record) bool {
	if r.Magic != other.Magic || r.Version != other.Version || r.Order != other.Order {
		return false
	}
	if r.Key != other.Key || r.HasArrhenius != other.HasArrhenius {
		return false
	}
	pairs := [][2]float64{
		{r.K, other.K},
		{r.K2, other.K2},
		{r.Time, other.Time},
		{r.Volume, other.Volume},
		{r.Conversion, other.Conversion},
		{r.ArrA, other.ArrA},
		{r.ArrEa, other.ArrEa},
		{r.ArrT, other.ArrT},
		{r.Selectivity, other.Selectivity},
		{r.SeriesPeakTime, other.SeriesPeakTime},
	}
	for _, p := range pairs {
		scale := math.Max(1, math.Max(math.Abs(p[0]), math.Abs(p[1])))
		if math.Abs(p[0]-p[1]) > Tol*scale {
			return false
		}
	}
	if len(r.Concentrations) != len(other.Concentrations) {
		return false
	}
	for name, c := range r.Concentrations {
		got := other.Concentrations[name]
		scale := math.Max(1, math.Max(math.Abs(c), math.Abs(got)))
		if math.Abs(got-c) > Tol*scale {
			return false
		}
	}
	return true
}

func cloneConfig(cfg reactor.Config) reactor.Config {
	out := cfg
	out.Initial = copyMap(cfg.Initial)
	out.Stoich = copyMap(cfg.Stoich)
	if cfg.Rate.K != nil {
		k := *cfg.Rate.K
		out.Rate.K = &k
	}
	if cfg.Rate.K1 != nil {
		k := *cfg.Rate.K1
		out.Rate.K1 = &k
	}
	if cfg.Rate.K2 != nil {
		k := *cfg.Rate.K2
		out.Rate.K2 = &k
	}
	if cfg.Rate.Arrhenius != nil {
		a := *cfg.Rate.Arrhenius
		out.Rate.Arrhenius = &a
	}
	return out
}

func copyMap(in map[string]float64) map[string]float64 {
	if in == nil {
		return nil
	}
	out := make(map[string]float64, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
