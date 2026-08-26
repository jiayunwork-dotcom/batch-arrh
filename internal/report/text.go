package report

import (
	"fmt"
	"strings"

	"batch-arrh/internal/reactor"
)

func renderTable(res *reactor.Result, rows int) string {
	if len(res.Trajectory) == 0 {
		return "(empty trajectory)\n"
	}
	names := res.SpeciesOrder()
	header := "    t (min)    X"
	for _, name := range names {
		header += fmt.Sprintf(" %10s", "C_"+name)
	}
	line := func(p reactor.Point) string {
		if p.C != nil {
			p.C[res.Key] = p.X
		}
		row := fmt.Sprintf("%12.4f %7.5f", p.Time, p.X)
		for _, name := range names {
			row += fmt.Sprintf(" %10.4g", p.C[name])
		}
		return row
	}
	width := len(header)
	if rows <= 0 {
		rows = 10
	}
	var b strings.Builder
	b.WriteString(header)
	b.WriteString("\n")
	if len(res.Trajectory) <= 2*rows {
		for _, p := range res.Trajectory {
			b.WriteString(line(p))
			b.WriteString("\n")
		}
		return b.String()
	}
	for i := 0; i < rows; i++ {
		b.WriteString(line(res.Trajectory[i]))
		b.WriteString("\n")
	}
	b.WriteString(strings.Repeat("-", width))
	b.WriteString("\n")
	for i := len(res.Trajectory) - rows; i < len(res.Trajectory); i++ {
		b.WriteString(line(res.Trajectory[i]))
		b.WriteString("\n")
	}
	return b.String()
}
