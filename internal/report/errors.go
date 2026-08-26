package report

import (
	"errors"

	"batch-arrh/internal/reactor"
)

var ErrNilResult = errors.New("cannot render a nil reactor result")

func requireResult(res *reactor.Result) error {
	if res == nil {
		return ErrNilResult
	}
	return nil
}

func ErrorLine(err error) string {
	if err == nil {
		return ""
	}
	return "batch-arrh: " + err.Error()
}
