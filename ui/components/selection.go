package components

import "sort"

// Selection is the set of checked items in a multi-select list, by key.
type Selection struct {
	keys map[string]bool
}

// NewSelection returns an empty selection.
func NewSelection() *Selection { return &Selection{keys: map[string]bool{}} }

// Has reports whether key is selected.
func (s *Selection) Has(key string) bool { return s.keys[key] }

// Set selects or deselects key.
func (s *Selection) Set(key string, on bool) {
	if on {
		s.keys[key] = true
	} else {
		delete(s.keys, key)
	}
}

// Toggle flips key.
func (s *Selection) Toggle(key string) { s.Set(key, !s.keys[key]) }

// SetAll selects or deselects every key given.
func (s *Selection) SetAll(keys []string, on bool) {
	for _, k := range keys {
		s.Set(k, on)
	}
}

// Clear deselects everything.
func (s *Selection) Clear() { s.keys = map[string]bool{} }

// Count is the number of selected keys.
func (s *Selection) Count() int { return len(s.keys) }

// Keys returns the selected keys, sorted.
func (s *Selection) Keys() []string {
	out := make([]string, 0, len(s.keys))
	for k := range s.keys {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// Retain drops selected keys that are no longer in the list, so a
// refresh never leaves invisible items selected.
func (s *Selection) Retain(present []string) {
	keep := make(map[string]bool, len(present))
	for _, k := range present {
		keep[k] = true
	}
	for k := range s.keys {
		if !keep[k] {
			delete(s.keys, k)
		}
	}
}

// AllOf reports whether every key given is selected (and there is one).
func (s *Selection) AllOf(keys []string) bool {
	for _, k := range keys {
		if !s.keys[k] {
			return false
		}
	}
	return len(keys) > 0
}
