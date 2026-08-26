package kinetics

import (
	"fmt"
	"sort"
	"strings"
)

type Species struct {
	Name string
	Conc float64
}

type SpeciesSet struct {
	names []string
	conc  map[string]float64
}

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

func NewEmptySpeciesSet() *SpeciesSet {
	return &SpeciesSet{conc: make(map[string]float64)}
}

func (s *SpeciesSet) Names() []string {
	out := make([]string, len(s.names))
	copy(out, s.names)
	return out
}

func (s *SpeciesSet) Get(name string) float64 {
	return s.conc[name]
}

func (s *SpeciesSet) Has(name string) bool {
	_, ok := s.conc[name]
	return ok
}

func (s *SpeciesSet) Set(name string, conc float64) {
	if !s.Has(name) {
		s.names = append(s.names, name)
	}
	s.conc[name] = conc
}

func (s *SpeciesSet) Len() int {
	return len(s.names)
}

func (s *SpeciesSet) Snapshot() map[string]float64 {
	out := make(map[string]float64, len(s.names))
	for name, c := range s.conc {
		out[name] = c
	}
	return out
}

func (s *SpeciesSet) Clone() *SpeciesSet {
	return NewSpeciesSet(s.Snapshot())
}

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
