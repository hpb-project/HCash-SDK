package utils

import (
	"github.com/hpb-project/HCash-SDK/core/utils/bn128"
	"math/big"
)

type BN128 struct {
	a    string
	b    string
	P    *big.Int
	Q    *big.Int
	n    *big.Int
	fq   bn128.Fq
	gRed bool
	g    bn128.G1
}

type Point struct {
	P [3]*big.Int
}

func NewPoint(bn *BN128, d1, d2 *big.Int) *Point {
	return &Point{
		P: [3]*big.Int{d1, d2, bn.fq.One()},
	}
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
	b.P = FIELD_MODULUS
	b.n = GROUP_MODULUS
	b.Q = b.n

	gX, _ := big.NewInt(int64(0)).SetString("77da99d806abd13c9f15ece5398525119d11e11e9836b2ee7d23f6159ad87d4", 16)
	gY, _ := big.NewInt(int64(0)).SetString("1485efa927f2ad41bff567eec88f32fb0a0f706588b4e41a8d587d008b7f875", 16)

	b.fq = bn128.NewFq(b.Q)
	gG1 := [2]*big.Int{gX, gY}
	b.g = bn128.NewG1(b.fq, gG1)

	return b
}

func (b *BN128) Zero() *Point {
	var p Point
	data := b.g.MulScalar(b.g.G, big.NewInt(0))
	p.P[0] = data[0]
	p.P[1] = data[1]
	p.P[2] = data[2]
	return &p
}

func (b *BN128) Rand() (*big.Int, error) {
	return b.fq.Rand()
}

func (b *BN128) Bytes(i *big.Int) string {
	return i.Text(16)
}

func (b *BN128) B_MAX() int {
	return B_MAX
}

func (b *BN128) Serialize(p *Point) (string, string) {
	if p.P[0] == nil && p.P[1] == nil {
		return "0000000000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000000000"
	}
	return b.Bytes(p.P[0]), b.Bytes(p.P[1])
}

func (b *BN128) UnSerialize(x, y string) *Point {
	if x == "0000000000000000000000000000000000000000000000000000000000000000" && y == "0000000000000000000000000000000000000000000000000000000000000000" {
		return b.Zero()
	} else {
		d1, _ := big.NewInt(0).SetString(x, 16)
		d2, _ := big.NewInt(0).SetString(y, 16)
		return NewPoint(b, d1, d2)
	}
}

func (b *BN128) Representation(p *Point) string {
	x, y := b.Serialize(p)
	return "0x" + x + y
}
