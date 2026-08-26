package thermo_test

import (
	"errors"
	"math"
	"testing"

	"batch-arrh/internal/thermo"
)

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

func TestArrheniusRejectsNegativeActivationEnergy(t *testing.T) {
	p := thermo.Arrhenius{A: 1e12, Ea: -100, T: 350}
	if err := thermo.ValidateArrhenius(p); err == nil {
		t.Fatal("ValidateArrhenius accepted Ea < 0, want error")
	}
}

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

func TestTemperatureToDoubleRate(t *testing.T) {
	p := thermo.Arrhenius{A: 1e12, Ea: 80000, T: 350}
	t2, err := thermo.TemperatureToScaleRate(p, 2)
	if err != nil {
		t.Fatalf("TemperatureToScaleRate: %v", err)
	}
	k1, err := p.Rate()
	if err != nil {
		t.Fatal(err)
	}
	k2, err := p.RateAt(t2)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(k2/k1-2) > 1e-9 {
		t.Errorf("k(T2)/k(T1) = %g, want 2", k2/k1)
	}
}

func TestLogDerivativeMatchesFiniteDifference(t *testing.T) {
	p := thermo.Arrhenius{A: 1e12, Ea: 80000, T: 350}
	exact, err := thermo.DlnKdT(p.Ea, p.T)
	if err != nil {
		t.Fatal(err)
	}
	fd, err := thermo.FiniteDifferenceDlnKdT(p, 0.01)
	if err != nil {
		t.Fatal(err)
	}
	rel := math.Abs(fd-exact) / exact
	if rel > 1e-4 {
		t.Errorf("d ln k / dT fd = %g, analytic = %g (rel %g)", fd, exact, rel)
	}
}

func TestQ10AgreesWithRateAt(t *testing.T) {
	p := thermo.Arrhenius{A: 1e12, Ea: 80000, T: 350}
	q, err := thermo.Q10(p)
	if err != nil {
		t.Fatal(err)
	}
	k1, _ := p.Rate()
	k2, _ := p.RateAt(360)
	want := k2 / k1
	if math.Abs(q-want) > 1e-12 {
		t.Errorf("Q10 = %g, want %g", q, want)
	}
}
