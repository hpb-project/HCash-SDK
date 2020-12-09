package ebigint

import (
	"github.com/hpb-project/HCash-SDK/core/utils/bn128"
	"math/big"
)

type NBigInt struct {
	*big.Int
	r  *Red
	fq bn128.Fq
}

func ToNBigInt(b *big.Int) *NBigInt {
	return &NBigInt{b, nil, nil}
}

func NewNBigInt(v int64) *NBigInt {
	return &NBigInt{big.NewInt(v), nil, nil}
}

func newRed(b *big.Int) *Red {
	return &Red{b}
}

func (this *NBigInt) Red(m *big.Int) *Red {
	return newRed(m)
}

func (this *NBigInt) ForceRed(r *Red) *NBigInt {
	this.r = r
	this.fq = bn128.NewFq(r.Int)
	return this
}

func (this *NBigInt) ToRed(r *Red) *NBigInt {
	return r.ConvertTo(this).ForceRed(r)
}

func (this *NBigInt) GetRed() *Red {
	return this.r
}

func (this *NBigInt) RedNeg() *NBigInt {
	t := &NBigInt{}
	t.Int = this.fq.Neg(this.Int)
	t.ForceRed(this.r)
	return t
}

func (this *NBigInt) RedMul(e *NBigInt) *NBigInt {
	t := &NBigInt{}
	t.Int = this.fq.Mul(this.Int, e.Int)
	t.ForceRed(this.r)
	return t
}

func (this *NBigInt) RedAdd(e *NBigInt) *NBigInt {
	t := &NBigInt{}
	t.Int = this.fq.Add(this.Int, e.Int)
	t.ForceRed(this.r)
	return t
}

func (this *NBigInt) RedSub(e *NBigInt) *NBigInt {
	t := &NBigInt{}
	t.Int = this.fq.Sub(this.Int, e.Int)
	t.ForceRed(this.r)
	return t
}

func (this *NBigInt) Exp(e *big.Int) *NBigInt {
	t := &NBigInt{}
	t.Int = this.fq.Exp(this.Int, e)
	t.ForceRed(this.r)
	return t
}
