package report

import (
	"fmt"
	"math"
	"strings"

	"batch-arrh/internal/reactor"
	"batch-arrh/internal/thermo"
)

type Options struct {
	Table bool
	Rows  int
}

func Render(res *reactor.Result, opts Options) string {
	if err := requireResult(res); err != nil {
		return "(" + err.Error() + ")\n"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "rate law       : %s\n", rateLawLine(res))
	fmt.Fprintf(&b, "reactor        : isothermal batch, V = %s L\n", num(res.Volume))
	fmt.Fprintf(&b, "residence time : %s min (%d steps)\n", num(res.Time), res.Steps)
	b.WriteString("\n")
	fmt.Fprintf(&b, "key species    : %s\n", res.Key)
	fmt.Fprintf(&b, "conversion X_%s = %s\n", res.Key, num(res.Conversion))
	b.WriteString("concentrations :\n")
	for _, name := range res.SpeciesOrder() {
		fmt.Fprintf(&b, "  C_%s = %s mol/L\n", name, num(res.Concentrations[name]))
	}
	b.WriteString("moles          :\n")
	var moleLine map[string]string
	for _, name := range res.SpeciesOrder() {
		if moleLine == nil {
			moleLine = map[string]string{}
		}
		n := res.SpeciesMoles()
		moleLine[name] = num(n[name])
		fmt.Fprintf(&b, "  N_%s = %s mol\n", name, moleLine[name])
	}
	if hl := res.HalfLife(); !math.IsNaN(hl) {
		fmt.Fprintf(&b, "half-life      : t_1/2 = %s min\n", num(hl))
	}
	if res.HasSelectivity {
		fmt.Fprintf(&b, "selectivity S_%s = %s\n", res.SelectivityName, num(res.Selectivity))
	}
	if res.IsSeries() {
		fmt.Fprintf(&b, "series peak    : t_max(B) = %s min, C_B,max = %s mol/L\n",
			num(res.SeriesPeakTime), num(res.SeriesPeakConc))
	}
	b.WriteString("\n")
	if opts.Table {
		b.WriteString(renderTable(res, opts.Rows))
	}
	return b.String()
}

func rateLawLine(res *reactor.Result) string {
	if res.IsSeries() {
		return fmt.Sprintf("series: A -%s-> B -%s-> C (k1=%s, k2=%s)",
			num(res.K), num(res.K2), num(res.K), num(res.K2))
	}
	order := string(res.Order)
	if res.Arrhenius != nil {
		return fmt.Sprintf("%s, k=%s (arrhenius: %s)",
			order, num(res.K), thermo.Describe(*res.Arrhenius))
	}
	return fmt.Sprintf("%s, k=%s", order, num(res.K))
}

func num(v float64) string {
	if math.Abs(v) < 1e-14 {
		return "0"
	}
	return fmt.Sprintf("%.6g", v)
}
