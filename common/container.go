package common

type Uint64Set struct {
	set map[uint64]struct{}
}

func NewUint64Set() *Uint64Set {
	return &Uint64Set{
		set: map[uint64]struct{}{},
	}
}

func (s *Uint64Set) Add(values ...uint64) *Uint64Set {
	for _, value := range values {
		s.set[value] = struct{}{}
	}
	return s
}

func (s *Uint64Set) IntersectOf(values ...uint64) *Uint64Set {
	new := NewUint64Set()
	for _, value := range values {
		if _, exist := s.set[value]; exist {
			new.Add(value)
		}
	}
	return new
}

func (s *Uint64Set) Intersect(values ...uint64) *Uint64Set {
	new := s.IntersectOf(values...)
	s.set = new.set
	return s
}

func (s *Uint64Set) Visit(visit func(uint64) bool) *Uint64Set {
	for v, _ := range s.set {
		if !visit(v) {
			break
		}
	}
	return s
}

func (s *Uint64Set) Len() int {
	return len(s.set)
}

type StringSet struct {
	set map[string]struct{}
}

func NewStringSet() *StringSet {
	return &StringSet{
		set: map[string]struct{}{},
	}
}

func (s *StringSet) Add(values ...string) *StringSet {
	for _, value := range values {
		s.set[value] = struct{}{}
	}
	return s
}

func (s *StringSet) Delete(values ...string) *StringSet {
	for _, value := range values {
		delete(s.set, value)
	}
	return s
}

func (s *StringSet) Visit(visit func(string) bool) *StringSet {
	for value, _ := range s.set {
		if !visit(value) {
			break
		}
	}
	return s
}

func (s *StringSet) In(values ...string) bool {
	for _, value := range values {
		if _, ok := s.set[value]; !ok {
			return false
		}
	}
	return true
}
