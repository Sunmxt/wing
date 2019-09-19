package common

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
