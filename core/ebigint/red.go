package ebigint

import "math/big"

type Red struct {
	*big.Int
}

func NewRed(b *big.Int) *Red {
	return &Red{b}
}

func (this *Red) ConvertTo(b *NBigInt) *NBigInt {
	var r = new(big.Int).Mod(b.Int, this.Int)
	return ToNBigInt(r)
}
