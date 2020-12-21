package core

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/hpb-project/HCash-SDK/core/ebigint"
)

type InnerProductProof struct {
	L []Point
	R []Point
	A *ebigint.NBigInt
	B *ebigint.NBigInt
}

func (i *InnerProductProof) Serialize() string {
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

func generateProof(base *GeneratorParams, P Point, as *FieldVector, bs *FieldVector,
	ls []Point, rs []Point, previousChallenge *ebigint.NBigInt) *InnerProductProof {
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

	var u = base.GetH()
	var L = gRight.Commit(asLeft).Add(hLeft.Commit(bsRight)).Add(u.Mul(cL))
	var R = gLeft.Commit(asRight).Add(hRight.Commit(bsLeft)).Add(u.Mul(cR))

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

	bytes, _ := arguments.Pack(
		b128.Bytes(previousChallenge.Int),
		[2]string(b128.Serialize(L)),
		[2]string(b128.Serialize(R)),
	)
	var x = Hash(hex.EncodeToString(bytes))
	var xInv = x.RedInvm()

	var gPrime = gLeft.Times(xInv).Add(gRight.Times(x))
	var hPrime = hLeft.Times(x).Add(hRight.Times(xInv))
	var aPrime = asLeft.Times(x).Add(asRight.Times(xInv))
	var bPrime = bsLeft.Times(xInv).Add(bsRight.Times(x))

	var PPrime = L.Mul(x.RedMul(x)).Add(R.Mul(xInv.RedMul(xInv))).Add(P)
	var basePrime = NewGeneratorParams(u, gPrime, hPrime)

	return generateProof(basePrime, PPrime, aPrime, bPrime, ls, rs, x)
}

type InnerProductProver struct {
}

type InnerProduct_witness struct {
	L *FieldVector
	R *FieldVector
}

type InnerProduct_statement struct {
	PrimeBase *GeneratorParams
	P         Point
}

func (t InnerProductProver) GenerateProof(statement InnerProduct_statement,
	witness InnerProduct_witness, salt *ebigint.NBigInt) *InnerProductProof {

	base := statement.PrimeBase
	P := statement.P
	l := witness.L
	r := witness.R
	return generateProof(base, P, l, r, []Point{}, []Point{}, salt)
}
