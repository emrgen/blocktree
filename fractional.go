package blocktree

import "fmt"

// ref: https://github.com/drifting-in-space/fractional_index/blob/main/src/fract_index.rs

const (
	terminator = uint8(128)
)

// FracIndex is a fractional index of a block.
type FracIndex struct {
	bytes []byte
}

// DefaultFracIndex creates a new fractional index.
func DefaultFracIndex() *FracIndex {
	return &FracIndex{bytes: []uint8{terminator}}
}

func fromBytes(bytes []byte) *FracIndex {
	return &FracIndex{bytes: bytes}
}

func fromUnterminated(bytes []byte) *FracIndex {
	return &FracIndex{bytes: append(bytes, terminator)}
}

func NewBefore(next *FracIndex) *FracIndex {
	return fromUnterminated(newBefore(next.bytes))
}

func NewBetween(left, right *FracIndex) (*FracIndex, error) {
	shorterLen := len(left.bytes)
	if len(right.bytes) < shorterLen {
		shorterLen = len(right.bytes)
	}

	shorterLen -= 1 // don't count the last byte, which may be the terminator

	for i := 0; i < shorterLen; i++ {
		if left.bytes[i] < right.bytes[i]-1 {
			bytes := make([]byte, i+1)
			copy(bytes, left.bytes)
			bytes[i] = (left.bytes[i] + right.bytes[i]) / 2
			return fromUnterminated(bytes), nil
		}

		if left.bytes[i] == right.bytes[i]-1 {
			bytes := make([]byte, len(left.bytes))
			copy(bytes, left.bytes[:i+1])
			copy(bytes[i+1:], newAfter(left.bytes[i+1:]))
			return fromUnterminated(bytes), nil
		}

		if left.bytes[i] > right.bytes[i] {
			return nil, fmt.Errorf("left index %v is not less than right index %v", left, right)
		}
	}

	// if we get here, the bytes are equal up to the length of the shorterLen
	if len(left.bytes) < len(right.bytes) {
		// if left and right are equal up to the length of the shorterLen
		// find the before FracIndex bytes from the right's suffix
		if uint8(right.bytes[shorterLen]) < terminator {
			return nil, fmt.Errorf("left index %v is not less than right index %v", left, right)
		}

		newSuffix := newBefore(right.bytes[shorterLen+1:])
		bytes := make([]byte, shorterLen+1+len(newSuffix))
		copy(bytes, right.bytes[:shorterLen+1])
		copy(bytes[shorterLen+1:], newSuffix)
		return fromUnterminated(bytes), nil
	} else if len(left.bytes) > len(right.bytes) {
		// if left and right are equal up to the length of the shorterLen
		// find the after FracIndex bytes from the left's suffix
		if uint8(left.bytes[shorterLen]) >= terminator {
			return nil, fmt.Errorf("left index %v is not less than right index %v", left, right)
		}

		newSuffix := newAfter(left.bytes[shorterLen+1:])
		bytes := make([]byte, shorterLen+1+len(newSuffix))
		copy(bytes, left.bytes[:shorterLen+1])
		copy(bytes[shorterLen+1:], newSuffix)
		return fromUnterminated(bytes), nil
	} else {
		return nil, fmt.Errorf("left index %v is equal to the right index %v", left, right)
	}
}

func NewAfter(prev *FracIndex) *FracIndex {
	return fromUnterminated(newAfter(prev.bytes))
}

func (f *FracIndex) Compare(other *FracIndex) int {
	for i := 0; i < len(f.bytes) && i < len(other.bytes); i++ {
		if f.bytes[i] < other.bytes[i] {
			return -1
		} else if f.bytes[i] > other.bytes[i] {
			return 1
		}
	}

	// if we get here, the bytes are equal up to the length of the shorter one
	if len(f.bytes) < len(other.bytes) {
		return -1
	}

	if len(f.bytes) > len(other.bytes) {
		return 1
	}

	return 0
}

func (f *FracIndex) Equals(other *FracIndex) bool {
	return f.Compare(other) == 0
}

func (f *FracIndex) String() string {
	ints := make([]int, len(f.bytes))
	for i, b := range f.bytes {
		ints[i] = int(b)
	}

	return fmt.Sprintf("FracIndex(%v)", ints)
}

func newBefore(index_bytes []byte) []byte {
	for i := 0; i < len(index_bytes); i++ {
		if uint8(index_bytes[i]) > terminator {
			bytes := make([]byte, i)
			copy(bytes, index_bytes)
			return bytes
		}

		if uint8(index_bytes[i]) > uint8(0) {
			bytes := make([]byte, i+1)
			copy(bytes, index_bytes)
			bytes[i] -= 1
			return bytes
		}
	}

	panic("should never reach the end of a properly-terminated fractional index without finding a byte greater than 0")
}

func newAfter(index_bytes []byte) []byte {
	for i := 0; i < len(index_bytes); i++ {
		if index_bytes[i] < terminator {
			ret := make([]byte, i)
			copy(ret, index_bytes)
			return ret
		}

		if index_bytes[i] < uint8(255) {
			ret := make([]byte, i+1)
			copy(ret, index_bytes)
			ret[i] += 1
			return ret
		}
	}

	panic("should never reach the end of a properly-terminated fractional index without finding a byte less than 255")
}
