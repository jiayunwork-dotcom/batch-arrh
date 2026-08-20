package thermo

// CelsiusToKelvin converts a temperature given on the celsius scale to the
// absolute kelvin scale. The constant-volume models are isothermal, so the
// converted value is evaluated once and reused by every rate-law call.
func CelsiusToKelvin(c float64) float64 {
	return c + KelvinAtZeroC
}

// KelvinToCelsius converts an absolute temperature back to celsius for
// display purposes.
func KelvinToCelsius(k float64) float64 {
	return k - KelvinAtZeroC
}

// KiloJoulesToJoules normalises an activation energy expressed in kJ/mol to
// the J/mol convention used with R.
func KiloJoulesToJoules(kj float64) float64 {
	return kj * JoulesPerKilojoule
}

// JoulesToKiloJoules is the inverse of KiloJoulesToJoules, used only for
// reporting.
func JoulesToKiloJoules(j float64) float64 {
	return j / JoulesPerKilojoule
}

// NormaliseTemperature returns t unchanged when it is already on the kelvin
// scale. The flag is kept explicit so callers state which convention they
// feed; there is no automatic guessing of units in the models.
func NormaliseTemperature(kelvin bool, t float64) float64 {
	if kelvin {
		return t
	}
	return CelsiusToKelvin(t)
}
