package ebigint

import "math/big"

type Red struct {
	m *big.Int
}

func (this *Red) ConvertTo(b *NBigInt) *NBigInt{
	var r = this.m.Mod(this.m, b.Int)
	return ToNBigInt(r)
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

