package thermo_test

import (
	"errors"
	"math"
	"testing"

	"batch-arrh/internal/thermo"
)

// TestArrheniusRateConstant checks that Rate() evaluates the Arrhenius law
// k = A*exp(-Ea/(R*T)) with the shared gas constant. With Ea = 0 the
// exponential collapses and k must equal A exactly.
func TestArrheniusRateConstant(t *testing.T) {
	t.Run("zero activation energy", func(t *testing.T) {
		p := thermo.Arrhenius{A: 2.0, Ea: 0, T: 300}
		got, err := p.Rate()
		if err != nil {
			t.Fatalf("Rate() returned error: %v", err)
		}
		const want = 2.0
		if got != want {
			t.Errorf("k = %g, want %g", got, want)
		}
	})
	t.Run("finite activation energy", func(t *testing.T) {
		p := thermo.Arrhenius{A: 1e12, Ea: 80000, T: 350}
		got, err := p.Rate()
		if err != nil {
			t.Fatalf("Rate() returned error: %v", err)
		}
		want := 1e12 * math.Exp(-80000/(thermo.R*350))
		rel := math.Abs(got-want) / want
		if rel > 1e-12 {
			t.Errorf("k = %g, want %g (rel err %g)", got, want, rel)
		}
	})
}

// TestArrheniusRateAtMatchesRate verifies that RateAt at the configured
// temperature agrees with Rate, so temperature overrides never change the
// base evaluation.
func TestArrheniusRateAtMatchesRate(t *testing.T) {
	p := thermo.Arrhenius{A: 5e11, Ea: 60000, T: 320}
	kBase, err := p.Rate()
	if err != nil {
		t.Fatalf("Rate() returned error: %v", err)
	}
	kAt, err := p.RateAt(320)
	if err != nil {
		t.Fatalf("RateAt() returned error: %v", err)
	}
	if math.Abs(kAt-kBase) > 1e-12*math.Max(1, kBase) {
		t.Errorf("RateAt(320) = %g, want %g from Rate()", kAt, kBase)
	}
}

// TestArrheniusRejectsZeroTemperature asserts that T = 0 is an error: the
// exponent Ea/(R*T) is undefined at the origin.
func TestArrheniusRejectsZeroTemperature(t *testing.T) {
	p := thermo.Arrhenius{A: 1e12, Ea: 80000, T: 0}
	_, err := p.Rate()
	if err == nil {
		t.Fatal("Rate() succeeded with T = 0, want error")
	}
	if !errors.Is(err, thermo.ErrNonPositiveTemperature) {
		t.Errorf("err = %v, want ErrNonPositiveTemperature", err)
	}
}

// TestArrheniusRejectsNegativeActivationEnergy asserts that a negative
// activation energy is rejected.
func TestArrheniusRejectsNegativeActivationEnergy(t *testing.T) {
	p := thermo.Arrhenius{A: 1e12, Ea: -100, T: 350}
	if err := thermo.ValidateArrhenius(p); err == nil {
		t.Fatal("ValidateArrhenius accepted Ea < 0, want error")
	}
}

// TestArrheniusRejectsNonPositivePreExponential asserts that A <= 0 is
// rejected, since a zero factor would silently disable every reaction.
func TestArrheniusRejectsNonPositivePreExponential(t *testing.T) {
	p := thermo.Arrhenius{A: -1, Ea: 80000, T: 350}
	if err := thermo.ValidateArrhenius(p); err == nil {
		t.Fatal("ValidateArrhenius accepted A < 0, want error")
	}
	p0 := thermo.Arrhenius{A: 0, Ea: 80000, T: 350}
	if err := thermo.ValidateArrhenius(p0); err == nil {
		t.Fatal("ValidateArrhenius accepted A = 0, want error")
	}
}

// TestArrheniusMonotonicInTemperature asserts that for a fixed A and Ea the
// rate constant grows with temperature, the property the "heating raises
// conversion" cross-rule relies on.
func TestArrheniusMonotonicInTemperature(t *testing.T) {
	p := thermo.Arrhenius{A: 1e12, Ea: 80000, T: 350}
	kHot, err := p.RateAt(420)
	if err != nil {
		t.Fatalf("RateAt(420) returned error: %v", err)
	}
	kCold, err := p.RateAt(350)
	if err != nil {
		t.Fatalf("RateAt(350) returned error: %v", err)
	}
	if kHot <= kCold {
		t.Errorf("k(420 K) = %g, want > k(350 K) = %g", kHot, kCold)
	}
}

// TestActivationEnergyFromPoints asserts that the two-point fit recovers
// the activation energy that generated two rate constants. The values are
// produced by the forward Arrhenius law, so a perfect round trip is
// expected within floating-point noise.
func TestActivationEnergyFromPoints(t *testing.T) {
	p := thermo.Arrhenius{A: 1e12, Ea: 80000, T: 350}
	k350, err := p.RateAt(350)
	if err != nil {
		t.Fatalf("RateAt(350) returned error: %v", err)
	}
	k420, err := p.RateAt(420)
	if err != nil {
		t.Fatalf("RateAt(420) returned error: %v", err)
	}
	got, err := thermo.ActivationEnergyFromPoints(k350, 350, k420, 420)
	if err != nil {
		t.Fatalf("ActivationEnergyFromPoints returned error: %v", err)
	}
	const want = 80000.0
	rel := math.Abs(got-want) / want
	if rel > 1e-9 {
		t.Errorf("Ea = %g, want %g (rel err %g)", got, want, rel)
	}
}

// TestTemperatureForRateRoundTrip asserts that TemperatureForRate inverts
// RateAt: the temperature that produces k(T0) is T0 itself.
func TestTemperatureForRateRoundTrip(t *testing.T) {
	p := thermo.Arrhenius{A: 1e12, Ea: 80000, T: 350}
	k, err := p.RateAt(380)
	if err != nil {
		t.Fatalf("RateAt(380) returned error: %v", err)
	}
	got, err := p.TemperatureForRate(k)
	if err != nil {
		t.Fatalf("TemperatureForRate returned error: %v", err)
	}
	if math.Abs(got-380) > 1e-9 {
		t.Errorf("T = %g, want 380", got)
	}
}

// TestPreExponentialFromPointsRoundTrip asserts that a measured rate
// constant and the known Ea reproduce the original pre-exponential factor.
func TestPreExponentialFromPointsRoundTrip(t *testing.T) {
	p := thermo.Arrhenius{A: 5e11, Ea: 60000, T: 320}
	k, err := p.Rate()
	if err != nil {
		t.Fatalf("Rate() returned error: %v", err)
	}
	got, err := thermo.PreExponentialFromPoints(k, 320, 60000)
	if err != nil {
		t.Fatalf("PreExponentialFromPoints returned error: %v", err)
	}
	rel := math.Abs(got-p.A) / p.A
	if rel > 1e-9 {
		t.Errorf("A = %g, want %g (rel err %g)", got, p.A, rel)
	}
}
