package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const firstOrderCase = `{
  "initial_concentration": {"A": 2.0},
  "stoichiometry": {"A": -1, "P": 1},
  "rate": {"order": "first", "k": 0.05},
  "residence_time": 100.0
}`

func TestHealth(t *testing.T) {
	srv := New(DefaultConfig())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("health code %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "ok") {
		t.Fatalf("health body %q", rec.Body.String())
	}
}

func TestIntegrateEndpoint(t *testing.T) {
	srv := New(DefaultConfig())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/integrate", strings.NewReader(firstOrderCase))
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("integrate code %d body %s", rec.Code, rec.Body.String())
	}
	var resp integrateResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Conversion <= 0.99 || resp.Conversion >= 1 {
		t.Fatalf("conversion %v, want ~0.993", resp.Conversion)
	}
	if resp.Key != "A" {
		t.Fatalf("key %q, want A", resp.Key)
	}
	if resp.K != 0.05 {
		t.Fatalf("k %v, want 0.05", resp.K)
	}
}

func TestIntegrateIllegalK(t *testing.T) {
	srv := New(DefaultConfig())
	raw := []byte(`{
	  "initial_concentration": {"A": 2.0},
	  "stoichiometry": {"A": -1, "P": 1},
	  "rate": {"order": "first", "k": -0.05},
	  "residence_time": 100.0
	}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/integrate", bytes.NewReader(raw))
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d body %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "error") {
		t.Fatalf("body should contain error: %q", rec.Body.String())
	}
}

func TestIntegrateInvalidJSON(t *testing.T) {
	srv := New(DefaultConfig())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/integrate", strings.NewReader(`{bad`))
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestIntegrateMethodNotAllowed(t *testing.T) {
	srv := New(DefaultConfig())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/integrate", nil)
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestVerifyEndpoint(t *testing.T) {
	srv := New(DefaultConfig())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/verify", strings.NewReader(firstOrderCase))
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("verify code %d body %s", rec.Code, rec.Body.String())
	}
	var resp verifyResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if !resp.OK {
		t.Fatal("first-order closed form must agree with the integrator")
	}
	if resp.Damkohler != 5 {
		t.Fatalf("Da = %v, want 5", resp.Damkohler)
	}
}

func TestVerifyIllegalTemperature(t *testing.T) {
	srv := New(DefaultConfig())
	raw := []byte(`{
	  "initial_concentration": {"A": 2.0},
	  "stoichiometry": {"A": -1, "P": 1},
	  "rate": {"order": "first", "arrhenius": {"A": 1e12, "Ea": 80000.0, "T": 0}},
	  "residence_time": 100.0
	}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/verify", bytes.NewReader(raw))
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d body %s", rec.Code, rec.Body.String())
	}
}
