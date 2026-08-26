# batch-arrh

`batch-arrh` evaluates conversion, species concentrations and (for A→B→C)
intermediate selectivity of an isothermal, constant-volume batch reactor.
It is delivered as a headless HTTP service and a small CLI; it serves no
web frontend.

The kernel applies a first-order, second-order or consecutive-series rate
law. The rate constant may be given directly or as an Arrhenius triplet
`k = A·exp(−Ea/(R·T))` with `R = 8.314 J/(mol·K)`. A solved case can be
written as a versioned JSON snapshot and read back. The model is a single
batch material balance; it does not schedule plant campaigns or compute
heat-exchange networks.

## Rate laws

Constant-volume relation for the key species:

    C_A = C_A0 · (1 − X)
    C_i = C_i0 − (ν_i/ν_A) · C_A0 · X

First order, `−r_A = k·C_A`:

    X = 1 − exp(−k·t)
    Damköhler Da = k·t,  X = 1 − exp(−Da)

Second order, equal initials `−r_A = k·C_A·C_B` with `C_A0 = C_B0`:

    1/C_A − 1/C_A0 = k·t
    X = Da / (1 + Da),  Da = k·C_A0·t

Second order with `C_A0 ≠ C_B0` and 1:1 stoichiometry (`M = C_B0/C_A0`):

    X = M · (1 − θ) / (M − θ)
    θ = exp((1 − M) · C_A0 · k · t)

Series `A → B → C` with `k1`, `k2`:

    C_A + C_B + C_C = C_A0  at every instant
    t_max(B) = ln(k1/k2) / (k1 − k2)   (k1 ≠ k2)

Illegal parameters (`k < 0`, `T ≤ 0`, `Ea < 0`, negative initial
concentration, `t < 0`, `k` and Arrhenius given together, missing rate)
are returned as errors. `t = 0` is legal and returns the initial state.

## HTTP API

With no arguments, or with the `serve` subcommand, the process listens on
`:8080`:

```
go run .
go run . serve
```

`GET /api/health` returns `{"status":"ok"}`.

`POST /api/integrate` integrates a scenario:

```
{
  "initial_concentration": {"A": 2.0},
  "stoichiometry": {"A": -1, "P": 1},
  "rate": {"order": "first", "k": 0.05},
  "residence_time": 100.0
}
```

`POST /api/verify` checks the same payload against the closed-form
conversion (Damköhler for first order / equal-second, analytic series
or unequal-second formula).

Illegal domain input returns HTTP 422. Malformed JSON returns HTTP 400.

## CLI usage

```
go build -o batch-arrh .
./batch-arrh                                              # HTTP on :8080
./batch-arrh serve                                        # same
./batch-arrh integrate example/first-order.json --t 100
./batch-arrh integrate example/second-order.json --t 100
./batch-arrh integrate example/series.json --table
./batch-arrh integrate example/first-order-arrhenius.json --json
```

For the shipped first-order example (`k = 0.05`, `t = 100`) the kernel
reports `X = 1 − exp(−5) ≈ 0.993262` and `C_A ≈ 0.013476 mol/L`.

## Building and testing

```
go test ./...
```

## Project layout

    main.go              CLI plus default HTTP listen
    internal/reactor     batch integration, Damköhler, series, moles
    internal/kinetics    rate laws, stoichiometry, extent
    internal/thermo      Arrhenius k(T), inverse, Q10
    internal/report      text and JSON rendering
    internal/api         HTTP /api/health, /api/integrate, /api/verify
    internal/snapshot    versioned JSON case round-trip
    example/             offline sample inputs

## License

MIT.
