package reactor

import (
	"fmt"
	"math"
)

type DesignTarget struct {
	Order      int
	K          float64
	CA0        float64
	Conversion float64
}

func (d DesignTarget) RequiredResidence() (float64, error) {
	switch d.Order {
	case 1:
		return FirstOrderTimeForConversion(d.K, d.Conversion)
	case 2:
		return SecondOrderEqualTimeForConversion(d.K, d.CA0, d.Conversion)
	default:
		return 0, fmt.Errorf("design: unsupported order %d", d.Order)
	}
}

func (d DesignTarget) DamkohlerAt(t float64) float64 {
	switch d.Order {
	case 1:
		return FirstOrderDamkohler(d.K, t)
	case 2:
		return SecondOrderEqualDamkohler(d.K, d.CA0, t)
	default:
		return math.NaN()
	}
}

func (d DesignTarget) ConversionAt(t float64) float64 {
	da := d.DamkohlerAt(t)
	switch d.Order {
	case 1:
		return ConversionFromFirstOrderDa(da)
	case 2:
		return ConversionFromSecondOrderEqualDa(da)
	default:
		return math.NaN()
	}
}
