package shared

// StringSet offers a nice api to deal with string lists
type StringSet struct {
	elements map[string]struct{}
}

func NewStringSet() *StringSet {
	return &StringSet{elements: make(map[string]struct{})}
}

func (s *StringSet) Add(st string) {
	s.elements[st] = struct{}{}
}

func (s *StringSet) Contains(st string) bool {
	_, ok := s.elements[st]
	return ok
}

func (s *StringSet) Slice() []string {
	out := make([]string, 0, len(s.elements))
	for st := range s.elements {
		out = append(out, st)
	}
	return out
}

func (s *StringSet) Len() int {
	return len(s.elements)
}
