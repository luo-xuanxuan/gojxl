package jxl

type Fields interface {
	VisitFields(visitor Visitor) error
}

const kDirect uint32 = 0x80000000

// Distribution of U32 values for one particular selector. Represents either a
// power of two-sized range, or a single value. A separate type ensures this is
// only passed to the U32Enc ctor.
type U32Distr struct {
	d uint32
}

func NewU32Distr(d uint32) U32Distr {
	return U32Distr{
		d: d,
	}
}

func (distr *U32Distr) IsDirect() bool {
	return (distr.d & kDirect) != 0
}

// Only call if IsDirect().
func (distr *U32Distr) Direct() uint32 {
	return distr.d & (kDirect - 1)
}

// Only call if !IsDirect().
func (distr *U32Distr) ExtraBits() uint {
	return uint((distr.d & 0x1F) + 1)
}

func (distr *U32Distr) Offset() uint32 {
	return (distr.d >> 5) & 0x3FFFFFF
}

func Val(value uint32) U32Distr {
	return NewU32Distr(value | kDirect)
}

// Value - `offset` will be signaled in `bits` extra bits.
func BitsOffset(bits, offset uint32) U32Distr {
	return NewU32Distr(((bits - 1) & 0x1F) + ((offset & 0x3FFFFFF) << 5))
}

// Value will be signaled in `bits` extra bits.
func Bits(bits uint32) U32Distr {
	return BitsOffset(bits, 0)
}

type U32Enc struct {
	d_ [4]U32Distr
}

func NewU32Enc(d0, d1, d2, d3 U32Distr) U32Enc {
	return U32Enc{
		d_: [4]U32Distr{d0, d1, d2, d3},
	}
}

// TODO: implement GetDistr
