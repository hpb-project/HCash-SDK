package prover

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/hpb-project/HCash-SDK/core/ebigint"
	"github.com/hpb-project/HCash-SDK/core/utils"
	"github.com/hpb-project/HCash-SDK/core/utils/bn128"
)

type InnerProductProof struct {
	L []utils.Point
	R []utils.Point
	A *ebigint.NBigInt
	B *ebigint.NBigInt
}

func (i *InnerProductProof) Serialize() string {
	b128 := utils.NewBN128()
	var result = "0x"
	for _, l := range i.L {
		result += b128.Representation(l)[2:]
	}
	for _, r := range i.R {
		result += b128.Representation(r)[2:]
	}

	result += b128.Bytes(i.A.Int)[2:]
	result += b128.Bytes(i.B.Int)[2:]
	return result
}

func generateProof(base *GeneratorParams, P utils.Point, as *FieldVector, bs *FieldVector,
	ls []utils.Point, rs []utils.Point, previousChallenge *ebigint.NBigInt) *InnerProductProof {
	var n = as.Length()
	if n == 1 {
		proof := &InnerProductProof{}
		proof.L = ls
		proof.R = rs
		proof.A = as.GetVector()[0]
		proof.B = bs.GetVector()[0]
		return proof
	}

	var nPrime = n / 2
	var asLeft = as.Slice(0, nPrime)
	var asRight = as.Slice(nPrime, as.Length())
	var bsLeft = bs.Slice(0, nPrime)
	var bsRight = bs.Slice(nPrime, bs.Length())

	var gLeft = base.GetGS().Slice(0, nPrime)
	var gRight = base.GetGS().Slice(nPrime, base.GetGS().Length())

	var hLeft = base.GetHS().Slice(0, nPrime)
	var hRight = base.GetHS().Slice(nPrime, base.GetHS().Length())

	var cL = asLeft.InnerProduct(bsRight)
	var cR = asRight.InnerProduct(bsLeft)
	b128 := utils.NewBN128()
	var u = base.GetH()

	t1 := b128.G1.MulScalar(u, cL.Int)
	t2 := hLeft.Commit(bsRight)
	t3 := gRight.Commit(asLeft)
	ta_1 := b128.G1.Add(t3, t2)
	L := b128.G1.Add(ta_1, t1)

	m1 := b128.G1.MulScalar(u, cR.Int)
	m2 := hRight.Commit(bsLeft)
	m3 := gLeft.Commit(asRight)
	ma_1 := b128.G1.Add(m3, m2)
	R := b128.G1.Add(ma_1, m1)

	ls = append(ls, L)
	rs = append(rs, R)

	bytes32_T, _ := abi.NewType("bytes32", "", nil)
	bytes32_2T, _ := abi.NewType("bytes32[2]", "", nil)

	arguments := abi.Arguments{
		{
			Type: bytes32_T,
		},
		{
			Type: bytes32_2T,
		},
		{
			Type: bytes32_2T,
		},
	}
	p2_1, p2_2 := b128.Serialize(L)
	p3_1, p3_2 := b128.Serialize(R)
	bytes, _ := arguments.Pack(
		b128.Bytes(previousChallenge.Int),
		[2]string{p2_1, p2_2},
		[2]string{p3_1, p3_2},
	)
	var x = utils.Hash(hex.EncodeToString(bytes))
	fq := bn128.NewFq(x.GetRed().Number())
	var xInv = ebigint.ToNBigInt(fq.Inverse(x.Int)).ToRed(x.GetRed())
	var gPrime = gLeft.Times(xInv).Add(gRight.Times(x))
	var hPrime = hLeft.Times(x).Add(hRight.Times(xInv))
	var aPrime = asLeft.Times(x).Add(asRight.Times(xInv))
	var bPrime = bsLeft.Times(xInv).Add(bsRight.Times(x))

	tm1 := fq.Mul(x.Int, x.Int)
	a_1 := b128.G1.MulScalar(L, tm1)

	tm2 := fq.Mul(xInv.Int, xInv.Int)
	a_2 := b128.G1.MulScalar(R, tm2)
	var PPrime = b128.G1.Add(b128.G1.Add(a_1, a_2), P)
	var basePrime = NewGeneratorParams(u, gPrime, hPrime)

	return generateProof(basePrime, PPrime, aPrime, bPrime, ls, rs, x)
}

type InnerProductProver struct {
}

func (t InnerProductProver) GenerateProof(statement map[string]interface{},
	witness map[string]interface{}, salt *ebigint.NBigInt) *InnerProductProof {

	base := statement["primeBase"].(*GeneratorParams)
	P := statement["P"].(utils.Point)
	l := witness["l"].(*FieldVector)
	r := witness["r"].(*FieldVector)
	return generateProof(base, P, l, r, []utils.Point{}, []utils.Point{}, salt)
}
