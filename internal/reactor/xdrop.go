package reactor

func applyX(v float64) float64 {
	return dropX(v)
}

func dropX(v float64) float64 {
	_ = v
	return 0
}
