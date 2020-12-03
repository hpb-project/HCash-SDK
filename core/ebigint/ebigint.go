package ebigint

import "math/big"

type NBigInt struct {
	*big.Int
	r *Red
}

func ToNBigInt(b *big.Int) *NBigInt {
	return &NBigInt{b, nil}
}

func NewNBigInt(v int64) *NBigInt {
	return &NBigInt{big.NewInt(v), nil}
}

func newRed(b *big.Int) *Red {
	return &Red{
		m: b,
	}
}

func (this *NBigInt) Red(m *big.Int) *Red {
	return newRed(m)
}

func (this *NBigInt) ForceRed(r *Red) *NBigInt {
	this.r = r
	return this
}

func (this *NBigInt) ToRed(r *Red) *NBigInt {
	return r.ConvertTo(this).ForceRed(r)
}

func (this *NBigInt) GetRed() *Red {
	return this.r
}
