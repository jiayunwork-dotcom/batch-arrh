package kinetics_test

import (
	"errors"
	"testing"

	"batch-arrh/internal/kinetics"
)

// TestFirstOrderRate asserts -r_A = k*C_A for a first-order step.
func TestFirstOrderRate(t *testing.T) {
	law := kinetics.FirstOrder{K: 0.1}
	got, err := law.Rate(2.0, 0)
	if err != nil {
		t.Fatalf("Rate() returned error: %v", err)
	}
	const want = 0.2
	if got != want {
		t.Errorf("-r_A = %g, want %g", got, want)
	}
}

// TestFirstOrderHalfLife asserts t_1/2 = ln(2)/k for the linear law.
func TestFirstOrderHalfLife(t *testing.T) {
	law := kinetics.FirstOrder{K: 0.1}
	got := law.HalfLife()
	const want = 6.931471805599453
	if diff := got - want; diff < -1e-12 || diff > 1e-12 {
		t.Errorf("t_1/2 = %g, want %g", got, want)
	}
}

// TestSecondOrderRate asserts -r_A = k*C_A*C_B for a bimolecular step.
func TestSecondOrderRate(t *testing.T) {
	law := kinetics.SecondOrder{K: 0.01}
	got, err := law.Rate(2.0, 3.0)
	if err != nil {
		t.Fatalf("Rate() returned error: %v", err)
	}
	const want = 0.06
	if got != want {
		t.Errorf("-r_A = %g, want %g", got, want)
	}
}

// TestSecondOrderHalfLife asserts t_1/2 = 1/(k*C_A0).
func TestSecondOrderHalfLife(t *testing.T) {
	law := kinetics.SecondOrder{K: 0.01}
	got := law.HalfLife(2.0)
	const want = 50.0
	if got != want {
		t.Errorf("t_1/2 = %g, want %g", got, want)
	}
}

// TestStoichConsumptionRatio asserts that a 1:2 stoichiometry consumes two
// moles of B per mole of A.
func TestStoichConsumptionRatio(t *testing.T) {
	s := kinetics.NewStoich(map[string]float64{"A": -1, "B": -2, "P": 1})
	got, err := s.ConsumptionRatio("A", "B")
	if err != nil {
		t.Fatalf("ConsumptionRatio returned error: %v", err)
	}
	const want = 2.0
	if got != want {
		t.Errorf("mol B per mol A = %g, want %g", got, want)
	}
}

// TestStoichProductionRatio asserts that a product is produced at the
// stoichiometric ratio per mole of reactant consumed.
func TestStoichProductionRatio(t *testing.T) {
	s := kinetics.NewStoich(map[string]float64{"A": -2, "P": 3})
	got, err := s.ProductionRatio("P", "A")
	if err != nil {
		t.Fatalf("ProductionRatio returned error: %v", err)
	}
	const want = 1.5
	if got != want {
		t.Errorf("mol P per mol A = %g, want %g", got, want)
	}
}

// TestStoichValidateRejectsNoReactant asserts that a map with only products
// is not a valid reaction and fails validation.
func TestStoichValidateRejectsNoReactant(t *testing.T) {
	s := kinetics.NewStoich(map[string]float64{"P": 1})
	err := s.Validate()
	if err == nil {
		t.Fatal("Validate accepted a reaction without reactants, want error")
	}
	if !errors.Is(err, kinetics.ErrNoReactant) {
		t.Errorf("err = %v, want ErrNoReactant", err)
	}
}

// TestStoichValidateRejectsNoProduct asserts that a map with only reactants
// is rejected.
func TestStoichValidateRejectsNoProduct(t *testing.T) {
	s := kinetics.NewStoich(map[string]float64{"A": -1})
	if err := s.Validate(); err == nil {
		t.Fatal("Validate accepted a reaction without products, want error")
	}
}

// TestStoichReactantsProducts asserts that Reactants and Products split the
// coefficient map by sign.
func TestStoichReactantsProducts(t *testing.T) {
	s := kinetics.NewStoich(map[string]float64{"A": -1, "B": -1, "P": 1, "Q": 2})
	reactants := s.Reactants()
	if len(reactants) != 2 || reactants[0] != "A" || reactants[1] != "B" {
		t.Errorf("Reactants() = %v, want [A B]", reactants)
	}
	products := s.Products()
	if len(products) != 2 || products[0] != "P" || products[1] != "Q" {
		t.Errorf("Products() = %v, want [P Q]", products)
	}
}

// TestSpeciesSetOrderIsStable asserts that a set built from a map keeps a
// deterministic order regardless of map iteration randomness.
func TestSpeciesSetOrderIsStable(t *testing.T) {
	set := kinetics.NewSpeciesSet(map[string]float64{"C": 1, "A": 2, "B": 3})
	got := set.Names()
	if len(got) != 3 || got[0] != "A" || got[1] != "B" || got[2] != "C" {
		t.Errorf("Names() = %v, want [A B C]", got)
	}
	if set.Get("A") != 2.0 {
		t.Errorf("Get(A) = %g, want 2", set.Get("A"))
	}
}

// TestSpeciesSetUpdateKeepsOrder asserts that Set on a new species appends
// and Set on an existing species only changes the concentration.
func TestSpeciesSetUpdateKeepsOrder(t *testing.T) {
	set := kinetics.NewSpeciesSet(map[string]float64{"A": 1})
	set.Set("B", 2)
	set.Set("A", 4)
	got := set.Names()
	if len(got) != 2 || got[0] != "A" || got[1] != "B" {
		t.Errorf("Names() = %v, want [A B]", got)
	}
	if set.Get("A") != 4.0 {
		t.Errorf("Get(A) = %g, want 4", set.Get("A"))
	}
}
