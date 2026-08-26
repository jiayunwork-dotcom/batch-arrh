package kinetics

type Order string

const (
	OrderFirst  Order = "first"
	OrderSecond Order = "second"
	OrderSeries Order = "series"
)

func (o Order) Valid() bool {
	switch o {
	case OrderFirst, OrderSecond, OrderSeries:
		return true
	default:
		return false
	}
}

type Law interface {
	Order() Order
	Rate(cA, cB float64) (float64, error)
	String() string
}

func NewLaw(order Order, k float64) Law {
	switch order {
	case OrderFirst:
		return FirstOrder{K: k}
	case OrderSecond:
		return SecondOrder{K: k}
	default:
		return FirstOrder{K: k}
	}
}
