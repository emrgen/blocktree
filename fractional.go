package blocktree

import (
	"bytes"
	"fmt"
)

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
	return &FracIndex{bytes: []byte{terminator}}
}

func FracIndexFromBytes(bytes []byte) *FracIndex {
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
			buf := make([]byte, i+1)
			copy(buf, left.bytes)
			buf[i] = (left.bytes[i] + right.bytes[i]) / 2
			return fromUnterminated(buf), nil
		}

		if left.bytes[i] == right.bytes[i]-1 {
			buf := make([]byte, len(left.bytes))
			copy(buf, left.bytes[:i+1])
			copy(buf[i+1:], newAfter(left.bytes[i+1:]))
			return fromUnterminated(buf), nil
		}

		if left.bytes[i] > right.bytes[i] {
			return nil, fmt.Errorf("left index %v is not less than right index %v", left, right)
		}
	}

	// if we get here, the bytes are equal up to the length of the shorterLen
	if len(left.bytes) < len(right.bytes) {
		// if left and right are equal up to the length of the shorterLen
		// find the before FracIndex buf from the right's suffix
		if uint8(right.bytes[shorterLen]) < terminator {
			return nil, fmt.Errorf("left index %v is not less than right index %v", left, right)
		}

		newSuffix := newBefore(right.bytes[shorterLen+1:])
		buf := make([]byte, shorterLen+1+len(newSuffix))
		copy(buf, right.bytes[:shorterLen+1])
		copy(buf[shorterLen+1:], newSuffix)
		return fromUnterminated(buf), nil
	} else if len(left.bytes) > len(right.bytes) {
		// if left and right are equal up to the length of the shorterLen
		// find the after FracIndex buf from the left's suffix
		if uint8(left.bytes[shorterLen]) >= terminator {
			return nil, fmt.Errorf("left index %v is not less than right index %v", left, right)
		}

		newSuffix := newAfter(left.bytes[shorterLen+1:])
		buf := make([]byte, shorterLen+1+len(newSuffix))
		copy(buf, left.bytes[:shorterLen+1])
		copy(buf[shorterLen+1:], newSuffix)
		return fromUnterminated(buf), nil
	} else {
		return nil, fmt.Errorf("left index %v is equal to the right index %v", left, right)
	}
}

func NewAfter(prev *FracIndex) *FracIndex {
	return fromUnterminated(newAfter(prev.bytes))
}

func (f *FracIndex) Bytes() []byte {
	return f.bytes
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

func (f *FracIndex) Clone() *FracIndex {
	return FracIndexFromBytes(bytes.Clone(f.bytes))
}

func (f *FracIndex) String() string {
	nums := make([]uint8, len(f.bytes))
	copy(nums, f.bytes)

	return fmt.Sprintf("FracIndex(%v)", nums)
}

func newBefore(indexBytes []byte) []byte {
	for i := 0; i < len(indexBytes); i++ {
		if indexBytes[i] > terminator {
			buf := make([]byte, i)
			copy(buf, indexBytes)
			return buf
		}

		if indexBytes[i] > uint8(0) {
			buf := make([]byte, i+1)
			copy(buf, indexBytes)
			buf[i] -= 1
			return buf
		}
	}

	panic("should never reach the end of a properly-terminated fractional index without finding a byte greater than 0")
}

func newAfter(indexBytes []byte) []byte {
	for i := 0; i < len(indexBytes); i++ {
		if indexBytes[i] < terminator {
			ret := make([]byte, i)
			copy(ret, indexBytes)
			return ret
		}

		if indexBytes[i] < uint8(255) {
			ret := make([]byte, i+1)
			copy(ret, indexBytes)
			ret[i] += 1
			return ret
		}
	}

	panic("should never reach the end of a properly-terminated fractional index without finding a byte less than 255")
}
