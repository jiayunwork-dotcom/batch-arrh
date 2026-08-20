// Package kinetics describes species, stoichiometry and the concentration
// level rate laws of an isothermal liquid-phase reaction. A rate law here
// only turns concentrations into -r_A (mol/(L*min)); the reactor package is
// responsible for joining that rate law with the material balance and the
// constant-volume conversion relation.
package kinetics

import (
	"fmt"
	"sort"
	"strings"
)

// Species is a single chemical species with a name and a molar
// concentration in mol/L.
type Species struct {
	Name string
	Conc float64
}

// SpeciesSet is an ordered collection of species. Order is preserved from
// the moment a set is built, so reports and tables render columns in a
// stable sequence instead of a nondeterministic map order.
type SpeciesSet struct {
	names []string
	conc  map[string]float64
}

// NewSpeciesSet builds an ordered set from a name-to-concentration map.
// Keys are sorted on construction, which makes the resulting order
// deterministic for any input.
func NewSpeciesSet(init map[string]float64) *SpeciesSet {
	s := &SpeciesSet{conc: make(map[string]float64, len(init))}
	s.names = make([]string, 0, len(init))
	for name := range init {
		s.names = append(s.names, name)
	}
	sort.Strings(s.names)
	for _, name := range s.names {
		s.conc[name] = init[name]
	}
	return s
}

// NewEmptySpeciesSet returns a set without any members, ready for Set calls.
func NewEmptySpeciesSet() *SpeciesSet {
	return &SpeciesSet{conc: make(map[string]float64)}
}

// Names returns the species names in the deterministic order of the set.
func (s *SpeciesSet) Names() []string {
	out := make([]string, len(s.names))
	copy(out, s.names)
	return out
}

// Get returns the concentration of a species, or zero when absent.
func (s *SpeciesSet) Get(name string) float64 {
	return s.conc[name]
}

// Has reports whether the species is a member of the set.
func (s *SpeciesSet) Has(name string) bool {
	_, ok := s.conc[name]
	return ok
}

// Set inserts or updates a species. New species are appended so the order
// stays stable across updates.
func (s *SpeciesSet) Set(name string, conc float64) {
	if !s.Has(name) {
		s.names = append(s.names, name)
	}
	s.conc[name] = conc
}

// Len returns the number of species in the set.
func (s *SpeciesSet) Len() int {
	return len(s.names)
}

// Snapshot copies the concentrations into a plain map.
func (s *SpeciesSet) Snapshot() map[string]float64 {
	out := make(map[string]float64, len(s.names))
	for name, c := range s.conc {
		out[name] = c
	}
	return out
}

// Clone returns a deep copy of the set.
func (s *SpeciesSet) Clone() *SpeciesSet {
	return NewSpeciesSet(s.Snapshot())
}

// String renders the set for diagnostics in a stable order.
func (s *SpeciesSet) String() string {
	var b strings.Builder
	b.WriteString("{")
	for i, name := range s.names {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "%s=%.6g", name, s.conc[name])
	}
	b.WriteString("}")
	return b.String()
}
