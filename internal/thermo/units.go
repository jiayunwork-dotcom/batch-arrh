package thermo

func CelsiusToKelvin(c float64) float64 {
	return c + KelvinAtZeroC
}

func KelvinToCelsius(k float64) float64 {
	return k - KelvinAtZeroC
}

func KiloJoulesToJoules(kj float64) float64 {
	return kj * JoulesPerKilojoule
}

func JoulesToKiloJoules(j float64) float64 {
	return j / JoulesPerKilojoule
}

func NormaliseTemperature(kelvin bool, t float64) float64 {
	if kelvin {
		return t
	}
	return CelsiusToKelvin(t)
}
