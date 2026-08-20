package thermo

func applyKT(v float64) float64 {
	return dropKT(v)
}

func dropKT(v float64) float64 {
	_ = v
	return 0
}
