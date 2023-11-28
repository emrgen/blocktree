package blocktree

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBeforeBytes(t *testing.T) {
	bytes := newBefore([]uint8{128})
	assert.Equal(t, []uint8{127}, bytes)

	bytes = newBefore([]uint8{127})
	assert.Equal(t, []uint8{126}, bytes)

	bytes = newBefore([]uint8{127, 128})
	assert.Equal(t, []uint8{126}, bytes)
}

func TestNewBeforeSimple(t *testing.T) {
	pos := DefaultFracIndex()
	if pos == nil {
		t.Error("NewFracIndex() returned nil")
		return
	}

	if pos.bytes == nil {
		t.Error("NewFracIndex() returned nil index")
	}

	assert.Equal(t, pos.bytes, []uint8{128})

	pos = NewBefore(pos)
	assert.Equal(t, pos.bytes, []uint8{127, 128})

	pos = NewBefore(pos)
	assert.Equal(t, pos.bytes, []uint8{126, 128})
}

func TestNewAfterSimple(t *testing.T) {
	pos := DefaultFracIndex()
	if pos == nil {
		t.Error("NewFracIndex() returned nil")
		return
	}

	if pos.bytes == nil {
		t.Error("NewFracIndex() returned nil index")
	}

	assert.Equal(t, pos.bytes, []uint8{128})

	pos = NewAfter(pos)
	assert.Equal(t, pos.bytes, []uint8{129, 128})

	pos = NewAfter(pos)
	assert.Equal(t, pos.bytes, []uint8{130, 128})
}

func TestNewBeforeLonger(t *testing.T) {
	pos := fromUnterminated([]uint8{100, 100, 3})
	assert.Equal(t, pos.bytes, []uint8{100, 100, 3, 128})

	pos = NewBefore(pos)
	assert.Equal(t, pos.bytes, []uint8{99, 128})

	pos = NewBefore(pos)
	assert.Equal(t, pos.bytes, []uint8{98, 128})
}

func TestNewAfterLonger(t *testing.T) {
	pos := fromUnterminated([]uint8{240, 240, 3})
	assert.Equal(t, pos.bytes, []uint8{240, 240, 3, 128})

	pos = NewAfter(pos)
	assert.Equal(t, pos.bytes, []uint8{241, 128})

	pos = NewAfter(pos)
	assert.Equal(t, pos.bytes, []uint8{242, 128})
}

func TestNewBeforeZeros(t *testing.T) {
	pos := fromUnterminated([]uint8{0, 0})
	assert.Equal(t, pos.bytes, []uint8{0, 0, 128})

	pos = NewBefore(pos)
	assert.Equal(t, pos.bytes, []uint8{0, 0, 127, 128})

	pos = NewBefore(pos)
	assert.Equal(t, pos.bytes, []uint8{0, 0, 126, 128})
}

func TestNewAfterMax(t *testing.T) {
	pos := fromUnterminated([]uint8{255, 255})
	assert.Equal(t, pos.bytes, []uint8{255, 255, 128})

	pos = NewAfter(pos)
	assert.Equal(t, pos.bytes, []uint8{255, 255, 129, 128})

	pos = NewAfter(pos)
	assert.Equal(t, pos.bytes, []uint8{255, 255, 130, 128})
}

// #[test]
// fn test_fractional_index() {
//     let mut indices: Vec<FractionalIndex> = Vec::new();

//     let c = FractionalIndex::default();

//     {
//         let mut m = c.clone();
//         let mut low = Vec::new();
//         for _ in 0..20 {
//             m = FractionalIndex::new_before(&m);
//             low.push(m.clone())
//         }

//         low.reverse();
//         indices.append(&mut low)
//     }

//     indices.push(c.clone());

//     {
//         let mut m = c.clone();
//         let mut high = Vec::new();
//         for _ in 0..20 {
//             m = FractionalIndex::new_after(&m);
//             high.push(m.clone())
//         }

//         indices.append(&mut high)
//     }

//     for i in 0..(indices.len() - 1) {
//         assert!(indices[i] < indices[i + 1])
//     }

//     for _ in 0..12 {
//         let mut new_indices: Vec<FractionalIndex> = Vec::new();
//         for i in 0..(indices.len() - 1) {
//             let cb = FractionalIndex::new_between(&indices[i], &indices[i + 1]).unwrap();
//             assert!(&indices[i] < &cb);
//             assert!(&cb < &indices[i + 1]);

//             let st = cb.to_string();
//             assert!(FractionalIndex::from_string(&st).unwrap() == cb);
//             assert!(st < indices[i + 1].to_string());

//             new_indices.push(cb);
//             new_indices.push(indices[i + 1].clone());
//         }

//         indices = new_indices;
//     }
// }

func TestBeforeWrap(t *testing.T) {
	pos := fromUnterminated([]uint8{0})
	assert.Equal(t, pos.bytes, []uint8{0, 128})

	pos = NewBefore(pos)
	assert.Equal(t, pos.bytes, []uint8{0, 127, 128})
}

func TestAfterWrap(t *testing.T) {
	pos := fromUnterminated([]uint8{255})
	assert.Equal(t, pos.bytes, []uint8{255, 128})

	pos = NewAfter(pos)
	assert.Equal(t, pos.bytes, []uint8{255, 129, 128})
}

func TestNewBetweenSimple(t *testing.T) {
	left := fromUnterminated([]uint8{100})
	right := fromUnterminated([]uint8{119})
	mid, err := NewBetween(left, right)
	assert.NoError(t, err)
	assert.Equal(t, mid.bytes, []uint8{109, 128})
}

func TestNewBetweenError(t *testing.T) {
	a := DefaultFracIndex()
	b := NewAfter(a)

	_, err := NewBetween(a, a)
	assert.Error(t, err)

	_, err = NewBetween(b, a)
	assert.Error(t, err)
}

func TestNewBetweenExtend(t *testing.T) {
	left := fromUnterminated([]uint8{100})
	right := fromUnterminated([]uint8{101})
	mid, err := NewBetween(left, right)
	if err != nil {
		t.Errorf("mid: %v", err)
		return
	}

	assert.Equal(t, mid.bytes, []uint8{100, 129, 128})
}

func TestNewBetweenPrefix(t *testing.T) {
	{
		left := fromUnterminated([]uint8{100})
		right := fromUnterminated([]uint8{100, 144})
		mid, err := NewBetween(left, right)
		assert.NoError(t, err)
		assert.Equal(t, []uint8{100, 144, 127, 128}, mid.bytes)
	}

	{
		left := fromUnterminated([]uint8{100, 122})
		right := fromUnterminated([]uint8{100})
		mid, err := NewBetween(left, right)
		assert.NoError(t, err)
		assert.Equal(t, mid.bytes, []uint8{100, 122, 129, 128})
	}

	{
		left := fromUnterminated([]uint8{100, 122})
		right := fromUnterminated([]uint8{100, 128})
		mid, err := NewBetween(left, right)
		assert.NoError(t, err)
		assert.Equal(t, mid.bytes, []uint8{100, 125, 128})
	}

	{
		left := fromUnterminated([]uint8{})
		right := fromUnterminated([]uint8{128, 192})
		mid, err := NewBetween(left, right)
		assert.NoError(t, err)
		assert.Equal(t, mid.bytes, []uint8{128, 128})
	}
}

func TestFractionalIndex(t *testing.T) {
	// not implemented
}
