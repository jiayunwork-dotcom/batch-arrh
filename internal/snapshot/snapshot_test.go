package snapshot

import (
	"os"
	"path/filepath"
	"testing"

	"batch-arrh/internal/reactor"
	"batch-arrh/internal/thermo"
)

func sampleFirstOrder() reactor.Config {
	cfg := reactor.DefaultConfig()
	cfg.Initial = map[string]float64{"A": 2.0}
	cfg.Stoich = map[string]float64{"A": -1, "P": 1}
	k := 0.05
	cfg.Rate.Order = "first"
	cfg.Rate.K = &k
	cfg.Time = 100
	return cfg
}

func TestSnapshotRoundTripAgrees(t *testing.T) {
	rec, err := Capture(sampleFirstOrder())
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "batch.snap.json")
	if err := WriteFile(path, rec); err != nil {
		t.Fatal(err)
	}
	got, err := ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Matches(rec) {
		t.Fatalf("round-trip mismatch: %+v vs %+v", got, rec)
	}
	if err := got.ReplayAgrees(); err != nil {
		t.Fatal(err)
	}
	if got.Conversion <= 0.99 || got.Conversion >= 1 {
		t.Fatalf("stored conversion %v out of expected range", got.Conversion)
	}
}

func TestSnapshotTruncationKeepsPriorFile(t *testing.T) {
	rec, err := Capture(sampleFirstOrder())
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	good := filepath.Join(dir, "good.json")
	if err := WriteFile(good, rec); err != nil {
		t.Fatal(err)
	}
	bad := filepath.Join(dir, "trunc.json")
	raw, err := os.ReadFile(good)
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) < 40 {
		t.Fatal("snapshot too small to truncate")
	}
	if err := os.WriteFile(bad, raw[:len(raw)/2], 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadFile(bad); err == nil {
		t.Fatal("truncated JSON must be rejected")
	}
	kept, err := ReadFile(good)
	if err != nil {
		t.Fatal(err)
	}
	if err := kept.ReplayAgrees(); err != nil {
		t.Fatal(err)
	}
	if !kept.Matches(rec) {
		t.Fatal("prior snapshot must still match the live kernel")
	}
}

func TestEmptyFileRejected(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.json")
	if err := os.WriteFile(path, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadFile(path); err == nil {
		t.Fatal("empty file must be rejected")
	}
}

func TestCaptureRejectsIllegalK(t *testing.T) {
	cfg := sampleFirstOrder()
	neg := -0.05
	cfg.Rate.K = &neg
	if _, err := Capture(cfg); err == nil {
		t.Fatal("negative k must not snapshot")
	}
}

func TestSeriesSnapshotReplayAgrees(t *testing.T) {
	cfg := reactor.DefaultConfig()
	cfg.Initial = map[string]float64{"A": 1.0}
	cfg.Stoich = map[string]float64{"A": -1, "B": 1, "C": 1}
	k1, k2 := 0.10, 0.02
	cfg.Rate.Order = "series"
	cfg.Rate.K1 = &k1
	cfg.Rate.K2 = &k2
	cfg.Time = 60
	rec, err := Capture(cfg)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "series.snap.json")
	if err := WriteFile(path, rec); err != nil {
		t.Fatal(err)
	}
	got, err := ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := got.ReplayAgrees(); err != nil {
		t.Fatal(err)
	}
	if got.SeriesPeakTime <= 0 {
		t.Fatalf("stored series peak time %v, want > 0", got.SeriesPeakTime)
	}
}

func TestTamperedConversionFailsReplay(t *testing.T) {
	rec, err := Capture(sampleFirstOrder())
	if err != nil {
		t.Fatal(err)
	}
	rec.Conversion = rec.Conversion * 0.5
	if err := rec.ReplayAgrees(); err == nil {
		t.Fatal("tampered conversion must fail replay")
	}
}

func TestArrheniusSnapshotStoresTriplet(t *testing.T) {
	cfg := reactor.DefaultConfig()
	cfg.Initial = map[string]float64{"A": 2.0}
	cfg.Stoich = map[string]float64{"A": -1, "P": 1}
	cfg.Rate.Order = "first"
	cfg.Rate.Arrhenius = &thermo.Arrhenius{A: 100, Ea: 20000, T: 300}
	cfg.Time = 20
	rec, err := Capture(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !rec.HasArrhenius || rec.ArrT != 300 {
		t.Fatalf("arrhenius triplet not stored: %+v", rec)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "arr.snap.json")
	if err := WriteFile(path, rec); err != nil {
		t.Fatal(err)
	}
	got, err := ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := got.ReplayAgrees(); err != nil {
		t.Fatal(err)
	}
}
