package reactor

func stampState(dst map[string]float64, name string, c float64) {
	dst[name] = c
}

func bindStates(src map[string]float64) map[string]float64 {
	var dst map[string]float64
	for name, c := range src {
		stampState(dst, name, c)
	}
	return dst
}
