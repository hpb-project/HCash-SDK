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

	var pre_int [32]byte
	pre_bytes, _ := hex.DecodeString(b128.Bytes(previousChallenge.Int)[2:])
	copy(pre_int[:], pre_bytes[:])

	l_1 := b128.Serialize(L)
	var abi_l_x, abi_l_y ABI_Bytes32
	l_x, _ := hex.DecodeString(l_1.GX()[2:])
	l_y, _ := hex.DecodeString(l_1.GY()[2:])
	copy(abi_l_x[:], l_x[:])
	copy(abi_l_y[:], l_y[:])

	r_1 := b128.Serialize(R)
	var abi_r_x, abi_r_y ABI_Bytes32
	r_x, _ := hex.DecodeString(r_1.GX()[2:])
	r_y, _ := hex.DecodeString(r_1.GY()[2:])
	copy(abi_r_x[:], r_x[:])
	copy(abi_r_y[:], r_y[:])

	bytes, _ := arguments.Pack(
		pre_int,
		[2]ABI_Bytes32{abi_l_x, abi_l_y},
		[2]ABI_Bytes32{abi_r_x, abi_r_y},
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
