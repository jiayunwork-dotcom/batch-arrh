package kinetics

// Order identifies the concentration level rate law that drives a reactor.
// The solver maps each order to a specific integration strategy: first
// order uses the closed form X = 1 - exp(-k*t), second order uses a
// Runge-Kutta integration of the conversion derivative, and series resolves
// the A -> B -> C network with explicit ODE integration.
type Order string

// Supported rate orders.
const (
	OrderFirst  Order = "first"
	OrderSecond Order = "second"
	OrderSeries Order = "series"
)

// Valid reports whether the order is implemented by the solver.
func (o Order) Valid() bool {
	switch o {
	case OrderFirst, OrderSecond, OrderSeries:
		return true
	default:
		return false
	}
}

// Law is the concentration level rate expression of a single reaction step.
// Rate returns -r_A in mol/(L*min) for the given concentrations of the key
// reactant (cA) and its co-reactant (cB, ignored by first order).
type Law interface {
	Order() Order
	Rate(cA, cB float64) (float64, error)
	String() string
}

// NewLaw builds the law matching a validated order and rate constant. The
// caller is responsible for checking k >= 0 before calling this.
func NewLaw(order Order, k float64) Law {
	switch order {
	case OrderFirst:
		return FirstOrder{K: k}
	case OrderSecond:
		return SecondOrder{K: k}
	default:
		// The switch is exhaustive over Valid orders; a caller that
		// reaches this branch has already been rejected upstream.
		return FirstOrder{K: k}
	}
}
