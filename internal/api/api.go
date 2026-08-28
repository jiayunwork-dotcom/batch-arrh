package api

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"

	"batch-arrh/internal/reactor"
)

type Server struct {
	mux  *http.ServeMux
	addr string
}

type Config struct {
	Addr string
}

type integrateResponse struct {
	Conversion     float64            `json:"conversion"`
	Key            string             `json:"key_species"`
	Concentrations map[string]float64 `json:"concentrations"`
	K              float64            `json:"rate_constant"`
	Order          string             `json:"rate_order"`
	Volume         float64            `json:"volume"`
	Time           float64            `json:"residence_time"`
	Selectivity    *float64           `json:"selectivity,omitempty"`
	SeriesPeakT    *float64           `json:"series_peak_time,omitempty"`
}

type verifyResponse struct {
	OK         bool    `json:"ok"`
	Conversion float64 `json:"conversion"`
	ClosedForm float64 `json:"closed_form"`
	Damkohler  float64 `json:"damkohler"`
	AbsDiff    float64 `json:"abs_diff"`
}

type requestError struct {
	code int
	msg  string
}

func (e *requestError) Error() string { return e.msg }

func DefaultConfig() Config {
	return Config{Addr: ":8080"}
}

func New(cfg Config) *Server {
	if cfg.Addr == "" {
		cfg.Addr = ":8080"
	}
	s := &Server{mux: http.NewServeMux(), addr: cfg.Addr}
	s.routes()
	return s
}

func (s *Server) Handler() http.Handler { return s.mux }

func (s *Server) Addr() string { return s.addr }

func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(s.addr, s.mux)
}

func ListenAndServe(cfg Config) error {
	return New(cfg).ListenAndServe()
}

func (s *Server) routes() {
	s.mux.HandleFunc("/api/health", s.handleHealth)
	s.mux.HandleFunc("/api/integrate", s.handleIntegrate)
	s.mux.HandleFunc("/api/verify", s.handleVerify)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) handleIntegrate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpError(w, http.StatusMethodNotAllowed, "POST only")
		return
	}
	res, err := solveBody(r)
	if err != nil {
		writeMapped(w, err)
		return
	}
	out := integrateResponse{
		Conversion:     res.Conversion,
		Key:            res.Key,
		Concentrations: res.Concentrations,
		K:              res.K,
		Order:          string(res.Order),
		Volume:         res.Volume,
		Time:           res.Time,
	}
	if res.HasSelectivity {
		v := res.Selectivity
		out.Selectivity = &v
	}
	if res.IsSeries() {
		p := res.SeriesPeakTime
		out.SeriesPeakT = &p
	}
	writeJSON(w, out)
}

func (s *Server) handleVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpError(w, http.StatusMethodNotAllowed, "POST only")
		return
	}
	res, err := solveBody(r)
	if err != nil {
		writeMapped(w, err)
		return
	}
	closed, da, err := closedFormOf(res)
	if err != nil {
		httpError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	diff := math.Abs(res.Conversion - closed)
	ok := diff <= 1e-5
	if !ok {
		httpError(w, http.StatusUnprocessableEntity, "closed-form conversion disagrees with the integrator")
		return
	}
	writeJSON(w, verifyResponse{
		OK:         true,
		Conversion: res.Conversion,
		ClosedForm: closed,
		Damkohler:  da,
		AbsDiff:    diff,
	})
}

func solveBody(r *http.Request) (reactor.Result, error) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return reactor.Result{}, &requestError{code: http.StatusBadRequest, msg: fmt.Sprintf("read body: %v", err)}
	}
	if len(body) == 0 {
		return reactor.Result{}, &requestError{code: http.StatusBadRequest, msg: "empty request body"}
	}
	cfg := reactor.DefaultConfig()
	if err := json.Unmarshal(body, &cfg); err != nil {
		return reactor.Result{}, &requestError{code: http.StatusBadRequest, msg: fmt.Sprintf("invalid JSON: %v", err)}
	}
	rx, err := reactor.New(&cfg)
	if err != nil {
		return reactor.Result{}, &requestError{code: http.StatusUnprocessableEntity, msg: err.Error()}
	}
	res, err := rx.Solve()
	if err != nil {
		return reactor.Result{}, &requestError{code: http.StatusUnprocessableEntity, msg: err.Error()}
	}
	return res, nil
}

func closedFormOf(res reactor.Result) (float64, float64, error) {
	cA0 := res.InitialKeyConcentration()
	switch {
	case res.IsSeries():
		pts, err := reactor.SolveSeries(cA0, res.K, res.K2, res.Time, res.Steps)
		if err != nil {
			return 0, 0, err
		}
		last := pts[len(pts)-1]
		x := reactor.ConversionFromConcentration(last.CA, cA0)
		return x, res.K * res.Time, nil
	case string(res.Order) == "first":
		da := reactor.FirstOrderDamkohler(res.K, res.Time)
		return reactor.ConversionFromFirstOrderDa(da), da, nil
	case string(res.Order) == "second":
		other := res.Concentrations[res.SecondOther]
		cB0 := other + cA0*res.Conversion
		if math.Abs(cB0-cA0) < 1e-9 {
			da := reactor.SecondOrderEqualDamkohler(res.K, cA0, res.Time)
			return reactor.ConversionFromSecondOrderEqualDa(da), da, nil
		}
		x := reactor.SecondOrderUnequalConversion(res.K, cA0, cB0, res.Time)
		da := res.K * cA0 * res.Time
		return x, da, nil
	default:
		return 0, 0, fmt.Errorf("unknown rate order %q", res.Order)
	}
}

func writeMapped(w http.ResponseWriter, err error) {
	if re, ok := err.(*requestError); ok {
		httpError(w, re.code, re.msg)
		return
	}
	httpError(w, http.StatusInternalServerError, err.Error())
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v)
}

func httpError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
