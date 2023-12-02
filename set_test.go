package blocktree

import "testing"

func TestNewSet(t *testing.T) {
	s := NewSet[int](1, 2, 3)
	if s.Cardinality() != 3 {
		t.Errorf("Expected cardinality 3, got %d", s.Cardinality())
	}
}

func TestSet_Add(t *testing.T) {
	s := NewSet[int](1, 2, 3)
	s.Add(4)
	if s.Cardinality() != 4 {
		t.Errorf("Expected cardinality 4, got %d", s.Cardinality())
	}
}

func TestSet_Remove(t *testing.T) {
	s := NewSet[int](1, 2, 3)
	s.Remove(3)
	if s.Cardinality() != 2 {
		t.Errorf("Expected cardinality 2, got %d", s.Cardinality())
	}
}

func TestSet_Contains(t *testing.T) {
	s := NewSet[int](1, 2, 3)
	if !s.Contains(3) {
		t.Errorf("Expected to contain 3")
	}
}

func TestSet_Cardinality(t *testing.T) {
	s := NewSet[int](1, 2, 3)
	if s.Cardinality() != 3 {
		t.Errorf("Expected cardinality 3, got %d", s.Cardinality())
	}
}

func TestSet_ToSlice(t *testing.T) {
	s := NewSet[int](1, 2, 3)
	slice := s.ToSlice()
	if len(slice) != 3 {
		t.Errorf("Expected slice length 3, got %d", len(slice))
	}
}

func TestSet_Union(t *testing.T) {
	s := NewSet[int](1, 2, 3)
	other := NewSet[int](3, 5, 6)
	union := s.Union(*other)
	if union.Cardinality() != 5 {
		t.Errorf("Expected cardinality 6, got %d", union.Cardinality())
	}
}

func TestSet_Intersect(t *testing.T) {
	s := NewSet[int](1, 2, 3)
	other := NewSet[int](2, 3, 4)
	intersection := s.Intersect(*other)
	if intersection.Cardinality() != 2 {
		t.Errorf("Expected cardinality 2, got %d", intersection.Cardinality())
	}
}

func TestSet_Difference(t *testing.T) {
	s := NewSet[int](1, 2, 3)
	other := NewSet[int](2, 3, 4)
	difference := s.Difference(other)
	if difference.Cardinality() != 1 {
		t.Errorf("Expected cardinality 1, got %d", difference.Cardinality())
	}
}

func TestSet_Equals(t *testing.T) {
	s := NewSet[int](1, 2, 3)
	other := NewSet[int](2, 3, 4)
	if s.Equals(other) {
		t.Errorf("Expected sets to be different")
	}
}

func TestSet_ForEach(t *testing.T) {
	s := NewSet[int](1, 2, 3)
	var sum int
	s.ForEach(func(i int) bool {
		sum += i
		return true
	})
	if sum != 6 {
		t.Errorf("Expected sum 6, got %d", sum)
	}
}
