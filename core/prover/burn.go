package prover

import (
	"encoding/hex"
	"github.com/hpb-project/HCash-SDK/core/types"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/hpb-project/HCash-SDK/core/ebigint"
	"github.com/hpb-project/HCash-SDK/core/utils"
)

type BurnProof struct {
	BA       utils.Point
	BS       utils.Point
	tCommits *GeneratorVector

	tHat *ebigint.NBigInt
	mu   *ebigint.NBigInt

	c     *ebigint.NBigInt
	s_sk  *ebigint.NBigInt
	s_b   *ebigint.NBigInt
	s_tau *ebigint.NBigInt

	ipProof *InnerProductProof
}

func (z BurnProof) Serialize() string {
	result := "0x"
	result += b128.Representation(z.BA)[2:]
	result += b128.Representation(z.BS)[2:]

	tcv := z.tCommits.GetVector()
	for _, commit := range tcv {
		result += b128.Representation(commit)[2:]
	}

	result += b128.Bytes(z.tHat.Int)[2:]
	result += b128.Bytes(z.mu.Int)[2:]
	result += b128.Bytes(z.c.Int)[2:]
	result += b128.Bytes(z.s_sk.Int)[2:]
	result += b128.Bytes(z.s_b.Int)[2:]
	result += b128.Bytes(z.s_tau.Int)[2:]

	result += z.ipProof.Serialize()[2:]

	return result
}

type BurnProver struct {
	params   *GeneratorParams
	ipProver *InnerProductProver
}

type BurnStatement struct {
	Input_CLn types.Publickey
	CLn       utils.Point
	Input_CRn types.Publickey
	CRn       utils.Point
	Input_y   types.Publickey
	Y         utils.Point
	Epoch     uint
	Sender    string
}

func NewBurnProver() BurnProver {
	params := NewGeneratorParams(int(32), nil, nil)
	return BurnProver{
		params:   params,
		ipProver: new(InnerProductProver),
	}
}

func (burn BurnProver) GenerateProof(statement BurnStatement, witness Witness) *BurnProof {
	var proof = &BurnProof{}

	bytes32_2T, _ := abi.NewType("bytes32[2]", "", nil)
	uint256_T, _ := abi.NewType("uint256", "", nil)
	address_T, _ := abi.NewType("address", "", nil)
	bytes32_T, _ := abi.NewType("bytes32", "", nil)
	arguments := abi.Arguments{
		{
			Type: bytes32_2T,
		},
		{
			Type: bytes32_2T,
		},
		{
			Type: bytes32_2T,
		},
		{
			Type: uint256_T,
		},
		{
			Type: address_T,
		},
	}
	vCLn := statement.Input_CLn //{{x1,y1}, {x2,y2}...}
	vCRn := statement.Input_CRn
	vy := statement.Input_y
	vepoch := statement.Epoch
	vsender := statement.Sender

	bytes, _ := arguments.Pack(
		vCLn,
		vCRn,
		vy,
		vepoch,
		vsender)

	var statementHash = utils.Hash(hex.EncodeToString(bytes))

	statement.CLn = b128.UnSerialize(vCLn)
	statement.CRn = b128.UnSerialize(vCRn)
	statement.Y = b128.UnSerialize(vy)

	witness.BDiff = ebigint.NewNBigInt(int64(witness.Input_bDiff)).ToRed(b128.Q())

	splits := strings.Split(witness.BDiff.Text(2), "")
	//println("len splits = ", len(splits), "xx ", splits)
	reversed := Reverse(splits)
	nArray := make([]*ebigint.NBigInt, len(reversed))
	for i, r := range reversed {
		n, _ := big.NewInt(0).SetString(r, 2)
		nArray[i] = ebigint.ToNBigInt(n).ToRed(b128.Q())
	}
	var aL = NewFieldVector(nArray)
	var aR = aL.Plus(ebigint.NewNBigInt(1).ToRed(b128.Q()).RedNeg())
	var alpha = b128.RanddomScalar()
	proof.BA = burn.params.Commit(alpha, aL, aR)

	var sL, sR *FieldVector
	{
		var vsL = make([]*ebigint.NBigInt, 32)
		var vsR = make([]*ebigint.NBigInt, 32)

		for i := 0; i < 32; i++ {
			vsL[i] = b128.RanddomScalar()
			vsR[i] = b128.RanddomScalar()
		}
		sL = NewFieldVector(vsL)
		sR = NewFieldVector(vsR)
	}
	var rho = b128.RanddomScalar()
	proof.BS = burn.params.Commit(rho, sL, sR)

	var y *ebigint.NBigInt
	{
		argumentsy := abi.Arguments{
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

		ybytes, _ := argumentsy.Pack(
			b128.Bytes(statementHash.Int),
			[2]string(b128.Serialize(proof.BA)),
			[2]string(b128.Serialize(proof.BS)),
		)

		y = utils.Hash(hex.EncodeToString(ybytes))
	}

	var vys = make([]*ebigint.NBigInt, 0)
	vys = append(vys, ebigint.NewNBigInt(1).ToRed(b128.Q()))
	for i := 1; i < 32; i++ {
		vys = append(vys, vys[i-1].RedMul(y))
	}
	ys := NewFieldVector(vys)
	z := utils.Hash(b128.Bytes(y.Int))

	var zs = make([]*ebigint.NBigInt, 0)
	zs = append(zs, z.RedExp(big.NewInt(2)))
	var twos = make([]*ebigint.NBigInt, 0)
	twos = append(twos, ebigint.NewNBigInt(1).ToRed(b128.Q()))
	for i := 1; i < 32; i++ {
		twos = append(twos, twos[i-1].RedMul(ebigint.NewNBigInt(2).ToRed(b128.Q())))
	}

	var twoTimesZs = NewFieldVector(twos).Times(zs[0])
	var lPoly = NewFieldVectorPolynomial(aL.Plus(z.RedNeg()), sL)
	var rPoly = NewFieldVectorPolynomial(ys.Hadamard(aR.Plus(z)).Add(twoTimesZs), sR.Hadamard(ys))
	var tPolyCoefficients = lPoly.InnerProduct(rPoly)

	var polyCommitment = NewPolyCommitment(*burn.params, tPolyCoefficients)
	proof.tCommits = NewGeneratorVector(polyCommitment.GetCommitments()) // just 2 of them

	argumentsx := abi.Arguments{
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
	var pcment = polyCommitment.GetCommitments()
	xbytes, _ := argumentsx.Pack(
		b128.Bytes(z.Int),
		[2]string(b128.Serialize(pcment[0])),
		[2]string(b128.Serialize(pcment[1])),
	)
	var x = utils.Hash(hex.EncodeToString(xbytes))

	var evalCommit = polyCommitment.Evaluate(x)
	proof.tHat = evalCommit.GetX()
	var tauX = evalCommit.GetR()
	proof.mu = alpha.RedAdd(rho.RedMul(x))

	var k_sk = b128.RanddomScalar()
	var k_b = b128.RanddomScalar()
	var k_tau = b128.RanddomScalar()

	var A_y = burn.params.GetG().Mul(k_sk)
	var A_b = burn.params.GetG().Mul(k_b).Add(statement.CRn.Mul(zs[0]).Mul(k_sk))
	var A_t = burn.params.GetG().Mul(k_b.RedNeg()).Add(burn.params.GetH().Mul(k_tau))
	var A_u = utils.GEpoch(statement.Epoch).Mul(k_sk)

	argumentsproofc := abi.Arguments{
		{
			Type: bytes32_T,
		},
		{
			Type: bytes32_2T,
		},
		{
			Type: bytes32_2T,
		},
		{
			Type: bytes32_2T,
		},
		{
			Type: bytes32_2T,
		},
	}
	cbytes, _ := argumentsproofc.Pack(
		b128.Bytes(x.Int),
		[2]string(b128.Serialize(A_y)),
		[2]string(b128.Serialize(A_b)),
		[2]string(b128.Serialize(A_t)),
		[2]string(b128.Serialize(A_u)),
	)
	proof.c = utils.Hash(hex.EncodeToString(cbytes))
	proof.s_sk = k_sk.RedAdd(proof.c.RedMul(witness.Input_sk))
	proof.s_b = k_b.RedAdd(proof.c.RedMul(witness.BDiff.RedMul(zs[0])))
	proof.s_tau = k_tau.RedAdd(proof.c.RedMul(tauX))

	var gs = burn.params.GetGS()
	var hPrimes = burn.params.GetHS().Hadamard(ys.Invert())
	var hExp = ys.Times(z).Add(twoTimesZs)

	var P = proof.BA.Add(proof.BS.Mul(x)).Add(gs.Sum().Mul(z.RedNeg())).Add(hPrimes.Commit(hExp))
	P = P.Add(burn.params.GetH().Mul(proof.mu.RedNeg())) // Statement P of protocol 1. should this be included in the calculation of v...?

	argumento := abi.Arguments{
		{
			Type: bytes32_T,
		},
	}
	obytes, _ := argumento.Pack(
		b128.Bytes(proof.c.Int),
	)
	var o = utils.Hash(hex.EncodeToString(obytes))
	var u_x = burn.params.GetG().Mul(o)
	P = P.Add(u_x.Mul(proof.tHat))
	var primeBase = NewGeneratorParams(u_x, gs, hPrimes)
	var ipStatement = InnerProduct_statement{}
	ipStatement.PrimeBase = primeBase
	ipStatement.P = P
	var ipWitness = InnerProduct_witness{}
	ipWitness.L = lPoly.Evaluate(x)
	ipWitness.R = rPoly.Evaluate(x)
	proof.ipProof = burn.ipProver.GenerateProof(ipStatement, ipWitness, o)

	return proof
}
