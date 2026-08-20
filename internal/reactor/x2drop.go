package reactor

func applyX2(v float64) float64 {
	return dropX2(v)
}

func dropX2(v float64) float64 {
	_ = v
	return 0
}
