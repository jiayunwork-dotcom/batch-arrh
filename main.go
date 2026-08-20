// Command batch-arrh computes conversion, concentrations and (optionally)
// selectivity of an isothermal batch reactor from a JSON scenario.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"batch-arrh/internal/reactor"
	"batch-arrh/internal/report"
)

const usage = `batch-arrh: isothermal batch reactor conversion accounting.

Reads a scenario JSON (initial concentrations, stoichiometry, rate constant
or Arrhenius triplet, residence time) and integrates the batch under a
constant volume, printing the key-species conversion X, every species
concentration and, for a series reaction, the intermediate selectivity and
peak time.

usage:
  batch-arrh integrate <scenario.json> [--t minutes] [--steps n] [--json] [--table] [--rows n]
  batch-arrh help

scenario.json fields:
  initial_concentration  map of species -> mol/L
  stoichiometry          map of species -> coefficient (reactants negative)
  rate                   {order: "first"|"second"|"series", k or arrhenius
                          (first/second), k1+k2 (series)}
  volume                 constant reactor volume in L (default 1.0)
  residence_time         batch time in minutes
  time_steps             RK4 intervals (default 2000)
  key_species            conversion basis (default "A")
  selectivity_species    intermediate used for selectivity (default "B")

Illegal parameters (k < 0, T <= 0, negative initial concentrations, t < 0,
ambiguous or missing rate data) are reported on stderr and exit non-zero.
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	switch os.Args[1] {
	case "integrate":
		if err := runIntegrate(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, report.ErrorLine(err))
			os.Exit(1)
		}
	case "help", "-h", "--help":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "batch-arrh: unknown command %q\n\n%s", os.Args[1], usage)
		os.Exit(2)
	}
}

// runIntegrate parses the flags, reads and validates the scenario, solves
// the reactor and prints the report. Every failure returns an error so the
// main function owns the exit code.
func runIntegrate(args []string) error {
	fs := flag.NewFlagSet("integrate", flag.ContinueOnError)
	timeArg := fs.Float64("t", 0, "override the residence time in minutes")
	stepsArg := fs.Int("steps", 0, "override the number of integration steps")
	asJSON := fs.Bool("json", false, "print machine-readable JSON")
	table := fs.Bool("table", false, "print the conversion trajectory")
	rowsArg := fs.Int("rows", 0, "print this many head and tail table rows")
	// The spec spells the invocation as `integrate <scenario.json> --t 100`,
	// with flags after the positional. The stdlib flag parser stops at the
	// first non-flag argument, so known flags are moved to the front first.
	if err := fs.Parse(reorderFlags(args)); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("integrate needs exactly one scenario JSON file")
	}
	data, err := os.ReadFile(fs.Arg(0))
	if err != nil {
		return fmt.Errorf("read scenario: %w", err)
	}
	cfg := reactor.DefaultConfig()
	if err := decodeJSON(data, &cfg); err != nil {
		return fmt.Errorf("parse scenario: %w", err)
	}
	// --t and --steps only override values that were actually passed on
	// the command line; flag.Visit walks exactly those.
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "t":
			cfg.Time = *timeArg
		case "steps":
			cfg.Steps = *stepsArg
		}
	})
	r, err := reactor.New(&cfg)
	if err != nil {
		return err
	}
	res, err := r.Solve()
	if err != nil {
		return err
	}
	if *asJSON {
		out, err := report.RenderJSON(&res)
		if err != nil {
			return err
		}
		fmt.Println(string(out))
		return nil
	}
	fmt.Print(report.Render(&res, report.Options{Table: *table, Rows: *rowsArg}))
	return nil
}

// decodeJSON unmarshals the scenario and rejects unknown fields so a typo
// in the input cannot silently change the batch.
func decodeJSON(data []byte, cfg *reactor.Config) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	return dec.Decode(cfg)
}

// reorderFlags moves the known flags (and their values) ahead of positional
// arguments. The stdlib flag parser stops at the first positional, so
// `integrate scenario.json --t 100` would otherwise leave --t unparsed.
// Unknown flags are left untouched and let flag.Parse report them.
func reorderFlags(args []string) []string {
	valueFlags := map[string]bool{"t": true, "steps": true, "rows": true}
	boolFlags := map[string]bool{"json": true, "table": true}
	flags := make([]string, 0, len(args))
	positionals := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") && arg != "-" && arg != "--" {
			name := strings.TrimLeft(arg, "-")
			key := name
			eq := strings.IndexByte(name, '=')
			if eq >= 0 {
				key = name[:eq]
			}
			switch {
			case valueFlags[key]:
				flags = append(flags, arg)
				if eq < 0 && i+1 < len(args) {
					i++
					flags = append(flags, args[i])
				}
			case boolFlags[key]:
				flags = append(flags, arg)
			default:
				// Unknown flag: let flag.Parse fail with its own message.
				return args
			}
			continue
		}
		positionals = append(positionals, arg)
	}
	return append(flags, positionals...)
}
