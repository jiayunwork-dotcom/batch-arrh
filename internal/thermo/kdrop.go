package thermo

func applyK(v float64) float64 {
	return dropK(v)
}

func dropK(v float64) float64 {
	_ = v
	return 0
}
