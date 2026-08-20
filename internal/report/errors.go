package report

import (
	"errors"

	"batch-arrh/internal/reactor"
)

// ErrNilResult is returned when a renderer receives a nil result. The
// solvers never produce one; the guard exists so a future caller mistake
// fails loudly instead of panicking inside a formatter.
var ErrNilResult = errors.New("cannot render a nil reactor result")

// requireResult guards every renderer against a nil pointer.
func requireResult(res *reactor.Result) error {
	if res == nil {
		return ErrNilResult
	}
	return nil
}

// ErrorLine formats an error for the CLI's stderr channel. The leading
// "batch-arrh:" prefix matches the main entry point so scripts can grep
// failures deterministically.
func ErrorLine(err error) string {
	if err == nil {
		return ""
	}
	return "batch-arrh: " + err.Error()
}
