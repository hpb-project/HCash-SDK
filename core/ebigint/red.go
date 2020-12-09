package ebigint

import "math/big"

type Red struct {
	*big.Int
}

func (this *Red) ConvertTo(b *NBigInt) *NBigInt {
	var r = this.Mod(this.Int, b.Int)
	return ToNBigInt(r)
}

func (this *Red) Number() *big.Int {
	return this.Int
}

//func (this *Red) IMod(b *NBigInt) *NBigInt {
//	return ToNBigInt(b.Mod(b.Int, this.m)).ForceRed(this)
//}

//func (this *Red) Add(a, b *NBigInt) *NBigInt {
//	if a.r != nil && a.r == b.r {
//		a.Add(a.Int,b.Int)
//	}
//}
//
