package utils

import (
	"github.com/hpb-project/HCash-SDK/core/ebigint"
	"github.com/hpb-project/HCash-SDK/core/utils/bn128"
	"math/big"
)

type BN128 struct {
	a    string
	b    string
	p    *big.Int
	n    *big.Int
	fq   bn128.Fq
	gRed bool
	G1   bn128.G1
}

type Point [3]*big.Int

func NewPoint(bn *BN128, d1, d2 *ebigint.NBigInt) Point {
	return Point{d1.Int, d2.Int, bn.fq.One()}
}

var (
	FIELD_MODULUS, _ = new(big.Int).SetString("30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47", 16)
	GROUP_MODULUS, _ = new(big.Int).SetString("30644e72e131a029b85045b68181585d2833e84879b9709143e1f593f0000001", 16)
	B_MAX            = 4294967295
)

func NewBN128() *BN128 {
	b := new(BN128)
	b.a = "0"
	b.b = "3"
	b.p = FIELD_MODULUS
	b.n = GROUP_MODULUS

	b.gRed = false

	gX, _ := big.NewInt(int64(0)).SetString("77da99d806abd13c9f15ece5398525119d11e11e9836b2ee7d23f6159ad87d4", 16)
	gY, _ := big.NewInt(int64(0)).SetString("1485efa927f2ad41bff567eec88f32fb0a0f706588b4e41a8d587d008b7f875", 16)

	b.fq = bn128.NewFq(b.n)
	gG1 := [2]*big.Int{gX, gY}
	b.G1 = bn128.NewG1(b.fq, gG1)

	return b
}

func (b *BN128) Zero() Point {
	data := b.G1.MulScalar(b.G1.G, big.NewInt(0))
	return Point{data[0], data[1], data[2]}
}

func (b *BN128) P() *ebigint.Red {
	return ebigint.ToNBigInt(big.NewInt(0)).Red(b.p)
}

func (b *BN128) Q() *ebigint.Red {
	return ebigint.ToNBigInt(big.NewInt(0)).Red(b.n)
}

func (b *BN128) FQ() bn128.Fq {
	return b.fq
}

func (b *BN128) RanddomScalar() *ebigint.NBigInt {

	r, _ := b.fq.Rand()
	nr := ebigint.ToNBigInt(r)
	return nr.ToRed(b.Q())
}

func (b *BN128) Bytes(i *big.Int) string {
	return "0x" + i.Text(16)
}

func (b *BN128) B_MAX() int {
	return B_MAX
}

func (b *BN128) Serialize(p Point) (string, string) {
	if p[0] == nil && p[1] == nil {
		return "0x0000000000000000000000000000000000000000000000000000000000000000", "0x0000000000000000000000000000000000000000000000000000000000000000"
	}
	return b.Bytes(p[0]), b.Bytes(p[1])
}

func (b *BN128) UnSerialize(x, y string) Point {
	if x == "0000000000000000000000000000000000000000000000000000000000000000" && y == "0000000000000000000000000000000000000000000000000000000000000000" {
		return b.Zero()
	} else {
		d1, _ := big.NewInt(0).SetString(x[2:], 16)
		d2, _ := big.NewInt(0).SetString(y[2:], 16)
		return NewPoint(b, ebigint.ToNBigInt(d1), ebigint.ToNBigInt(d2))
	}
}

func (b *BN128) Representation(p Point) string {
	x, y := b.Serialize(p)
	return x + y[2:]
}
