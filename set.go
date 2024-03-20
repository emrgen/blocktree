package blocktree

// Set is a set data structure
type Set[T comparable] struct {
	items map[T]bool
}

// NewSet creates a new set with the given entries
func NewSet[T comparable](entries ...T) *Set[T] {
	items := make(map[T]bool)
	for _, item := range entries {
		items[item] = true
	}

	return &Set[T]{
		items: items,
	}
}

// Remove removes an item from the set
func (s Set[T]) Remove(item T) {
	delete(s.items, item)
}

// Add adds an item to the set
func (s Set[T]) Add(item T) {
	s.items[item] = true
}

// Contains returns true if the set contains the item
func (s Set[T]) Contains(item T) bool {
	_, ok := s.items[item]
	return ok
}

// Cardinality returns the number of items in the set
func (s Set[T]) Cardinality() int {
	return len(s.items)
}

// ToSlice returns a slice of all items in the set
func (s Set[T]) ToSlice() []T {
	var slice []T
	for k := range s.items {
		slice = append(slice, k)
	}
	return slice
}

func (s Set[T]) Union(other Set[T]) *Set[T] {
	union := NewSet[T]()
	for k := range s.items {
		union.Add(k)
	}
	for k := range other.items {
		union.Add(k)
	}
	return union
}

func (s Set[T]) Intersect(other Set[T]) *Set[T] {
	intersection := NewSet[T]()
	for k := range s.items {
		if other.Contains(k) {
			intersection.Add(k)
		}
	}
	return intersection
}

func (s Set[T]) Difference(other *Set[T]) *Set[T] {
	difference := NewSet[T]()
	for k := range s.items {
		if !other.Contains(k) {
			difference.Add(k)
		}
	}
	return difference
}

func (s Set[T]) ForEach(cb func(T) bool) {
	for k := range s.items {
		if !cb(k) {
			break
		}
	}
}

func (s Set[T]) Equals(other *Set[T]) bool {
	if s.Cardinality() != other.Cardinality() {
		return false
	}

	for k := range s.items {
		if !other.Contains(k) {
			return false
		}
	}
	return true
}
