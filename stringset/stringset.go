package stringset

type Set map[string]struct{}

func New() Set {
	return Set{}
}

func NewFromSlice(slice []string) Set {
	s := New()
	for _, v := range slice {
		s.Add(v)
	}
	return s
}

func (s Set) Slice() []string {
	ret := []string{}
	for k := range s {
		ret = append(ret, k)
	}
	return ret
}

func (s Set) Add(v string) {
	s[v] = struct{}{}
}

func (s Set) Has(v string) bool {
	if _, exists := s[v]; exists {
		return true
	}
	return false
}

func (s Set) Remove(v string) {
	delete(s, v)
}

func (s Set) Copy() Set {
	copied := New()
	for v := range s {
		copied.Add(v)
	}
	return copied
}

// Join joins two sets and returns a function to restore its original state.
// Example: restore := s.Join(ss); /* some operation */ restore(s)
func (s Set) Join(ss Set) func(Set) {
	added := []string{}
	for v := range ss {
		if _, exists := s[v]; !exists {
			added = append(added, v)
			s.Add(v)
		}
	}

	return func(s Set) {
		for _, v := range added {
			s.Remove(v)
		}
	}
}
