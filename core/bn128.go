package core

import (
	"bytes"
	"encoding/hex"
	"github.com/hpb-project/HCash-SDK/common/types"
	"github.com/hpb-project/HCash-SDK/core/bn256"
	"github.com/hpb-project/HCash-SDK/core/ebigint"
	"log"
	"math/big"
)

var (
	FIELD_MODULUS, _ = new(big.Int).SetString("30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47", 16)
	GROUP_MODULUS, _ = new(big.Int).SetString("30644e72e131a029b85045b68181585d2833e84879b9709143e1f593f0000001", 16)
	bigzero          = "0x0000000000000000000000000000000000000000000000000000000000000000"
	B_MAX            = 4294967295
)

type BN128 struct {
	p  *big.Int
	n  *big.Int
	G1 *bn256.G1
}

type Point struct {
	p *bn256.G1
}

func newPoint(p *bn256.G1) Point {
	n := new(Point)
	n.p = p
	return *n
}

func (p Point) Mul(o *ebigint.NBigInt) Point {
	np := new(bn256.G1).ScalarMult(p.p, o.Int)
	return newPoint(np)
}

func (p Point) Add(o Point) Point {
	np := new(bn256.G1).Add(p.p, o.p)
	return newPoint(np)
}

func (p Point) Equal(o Point) bool {
	d1 := p.p.Marshal()
	d2 := o.p.Marshal()
	return bytes.Compare(d1, d2) == 0
}

func (p Point) Neg() Point {
	np := new(bn256.G1).Neg(p.p)
	return newPoint(np)
}

func (p Point) XY() (*big.Int, *big.Int) {
	if p.p != nil {
		data := p.p.Marshal()
		x := new(big.Int).SetBytes(data[0:32])
		y := new(big.Int).SetBytes(data[32:64])
		return x, y
	} else {
		return nil, nil
	}
}

func (p Point) String() string {
	data := p.p.Marshal()
	x := data[0:32]
	y := data[32:]
	up := types.Point{hex.EncodeToString(x), hex.EncodeToString(y)}
	return up.String()
}

func NewPoint(d1, d2 *big.Int) Point {
	x := BytePadding(d1.Bytes(), 32)
	y := BytePadding(d2.Bytes(), 32)
	m := BytesCombine(x, y)
	g, _ := new(bn256.G1).Unmarshal(m)
	return Point{g}
}

func BytePadding(data []byte, length int) []byte {
	datalen := len(data)
	ret := make([]byte, length)
	if datalen < length {
		copy(ret[length-datalen:], data)
	} else {
		copy(ret[:], data[datalen-length:])
	}
	return ret
}

func BytesCombine(pBytes ...[]byte) []byte {
	return bytes.Join(pBytes, []byte(""))
}

func NewBN128() *BN128 {
	b := new(BN128)
	b.p = FIELD_MODULUS
	b.n = GROUP_MODULUS

	gX, _ := hex.DecodeString("077da99d806abd13c9f15ece5398525119d11e11e9836b2ee7d23f6159ad87d4")
	gY, _ := hex.DecodeString("01485efa927f2ad41bff567eec88f32fb0a0f706588b4e41a8d587d008b7f875")
	m := BytesCombine(BytePadding(gX, 32), BytePadding(gY, 32))
	b.G1, _ = new(bn256.G1).Unmarshal(m)

	if b.G1 == nil {
		log.Println("new bn128 failed")
		return nil
	}

	return b
}

func (b *BN128) CurveG() Point {
	return newPoint(b.G1)
}

func (b *BN128) CurveRed() *ebigint.Red {
	return b.P()
}

func (b *BN128) Fq() bn256.Fq {
	return bn256.NewFq(b.p)
}

func (b *BN128) P() *ebigint.Red {
	return ebigint.NewRed(b.p)
}

func (b *BN128) Q() *ebigint.Red {
	return ebigint.NewRed(b.n)
}

func (b *BN128) Zero() Point {
	data := new(bn256.G1).ScalarMult(b.G1, big.NewInt(0))
	return Point{data}
}

func (b *BN128) RandomScalar() *ebigint.NBigInt {
	fq := bn256.NewFq(b.Q().Int)
	r, _ := fq.Rand()
	nr := ebigint.ToNBigInt(r).ForceRed(b.Q())
	return nr
}

func (b *BN128) Bytes(i *big.Int) string {
	hexstr := PaddingString(i.Text(16), 64)
	return "0x" + hexstr
}

func (b *BN128) B_MAX() int {
	return B_MAX
}

func (b *BN128) Serialize(p Point) types.Point {
	var x, y string
	gx, gy := p.XY()
	if gx == nil || gy == nil {
		x = bigzero
		y = bigzero
	} else {
		x = b.Bytes(gx)
		y = b.Bytes(gy)
	}
	return types.Point{x, y}
}

func (b *BN128) UnSerialize(pubkey types.Point) Point {
	x := pubkey.GX()
	y := pubkey.GY()
	if x == bigzero && y == bigzero {
		return b.Zero()
	} else {
		d1, _ := big.NewInt(0).SetString(x[2:], 16)
		d2, _ := big.NewInt(0).SetString(y[2:], 16)
		return NewPoint(d1, d2)
	}
}

func (b *BN128) Representation(p Point) string {
	key := b.Serialize(p)
	return "0x" + key.GX()[2:] + key.GY()[2:]
}
