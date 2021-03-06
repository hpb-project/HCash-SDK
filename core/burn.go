package core

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/hpb-project/HCash-SDK/common"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/hpb-project/HCash-SDK/core/ebigint"
)

type BurnProof struct {
	BA       Point
	BS       Point
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

func NewBurnProver() BurnProver {
	params := NewGeneratorParams(int(32), nil, nil)
	return BurnProver{
		params:   params,
		ipProver: new(InnerProductProver),
	}
}

type interBurnStatement struct {
	CLn    Point
	CRn    Point
	Y      Point
	Epoch  int
	Sender string
}

type interBurnWitness struct {
	bDiff *ebigint.NBigInt
	sk    *ebigint.NBigInt
}

func (burn BurnProver) tointerBurnStatement(istatement BurnStatement) (*interBurnStatement, error) {
	statement := &interBurnStatement{}
	statement.Epoch = istatement.Epoch
	statement.Sender = istatement.Sender

	statement.CLn = b128.UnSerialize(istatement.CLn)
	statement.CRn = b128.UnSerialize(istatement.CRn)
	statement.Y = b128.UnSerialize(istatement.Y)

	return statement, nil
}

func (burn BurnProver) tointerBurnWitness(iwitness BurnWitness) (*interBurnWitness, error) {
	witness := &interBurnWitness{}
	witness.bDiff = ebigint.NewNBigInt(int64(iwitness.BDiff)).ToRed(b128.Q())
	str_sk := iwitness.SK
	if strings.HasPrefix(str_sk, "0x") {
		str_sk = str_sk[2:]
	}
	sk, ok := big.NewInt(0).SetString(str_sk, 16)
	if !ok {
		return nil, errors.New("witness sk is invalid")
	} else {
		witness.sk = ebigint.ToNBigInt(sk).ForceRed(b128.Q())
	}
	return witness, nil
}

func (burn BurnProver) GenerateProof(istatement BurnStatement, iwitness BurnWitness) *BurnProof {
	var proof = &BurnProof{}
	var err error
	var statement *interBurnStatement
	var witness *interBurnWitness

	statement, err = burn.tointerBurnStatement(istatement)
	if err != nil {
		log.Printf("to inter burn statement failed, err:%s\n", err.Error())
		return nil
	}

	witness, err = burn.tointerBurnWitness(iwitness)
	if err != nil {
		log.Printf("to inter burn waitness failed, err:%s\n", err.Error())
		return nil
	}

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
	address := common.FromHex(istatement.Sender)
	var addr = ETH_ADDR{}
	copy(addr[:], address[:])

	epoch := new(big.Int).SetInt64(int64(istatement.Epoch))

	bytes, err := arguments.Pack(
		parsePoint2ABI_Bytes32_2(statement.CLn),
		parsePoint2ABI_Bytes32_2(statement.CRn),
		parsePoint2ABI_Bytes32_2(statement.Y),
		epoch,
		addr)
	//fmt.Println("statement addr = ", hex.EncodeToString(addr[:]))
	strbytes := hex.EncodeToString(bytes)

	var statementHash = Hash(strbytes)
	//fmt.Println("statementhash  = ", statementHash.Text(16))
	splits := strings.Split(witness.bDiff.Text(2), "")

	reversed := Reverse(splits)
	al_reversed := [32]string{""}
	copy(al_reversed[:], reversed)
	fmt.Println("al_re=", len(al_reversed))
	nArray := make([]*ebigint.NBigInt, len(al_reversed))
	for i := 0; i < 32; i++ {
		n := big.NewInt(0)
		n.SetString(al_reversed[i], 2)
		nArray[i] = ebigint.ToNBigInt(n).ToRed(b128.Q())
	}

	var aL = NewFieldVector(nArray)
	var aR = aL.Plus(ebigint.NewNBigInt(1).ToRed(b128.Q()).RedNeg())

	var alpha = b128.RandomScalar()
	proof.BA = burn.params.Commit(alpha, aL, aR)
	fmt.Println("ba=", proof.BA.String())

	var sL, sR *FieldVector
	{
		var vsL = make([]*ebigint.NBigInt, 32)
		var vsR = make([]*ebigint.NBigInt, 32)

		for i := 0; i < 32; i++ {
			vsL[i] = b128.RandomScalar()
			vsR[i] = b128.RandomScalar()
		}
		sL = NewFieldVector(vsL)
		sR = NewFieldVector(vsR)
	}

	var rho = b128.RandomScalar()
	proof.BS = burn.params.Commit(rho, sL, sR)
	fmt.Println("bs=", proof.BS.String())

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
			parseBigInt2ABI_Bytes32(statementHash),
			parsePoint2ABI_Bytes32_2(proof.BA),
			parsePoint2ABI_Bytes32_2(proof.BS),
		)
		y = Hash(hex.EncodeToString(ybytes))
	}

	fmt.Println("y=", y.String())

	var vys = make([]*ebigint.NBigInt, 0)
	vys = append(vys, ebigint.NewNBigInt(1).ToRed(b128.Q()))
	for i := 1; i < 32; i++ {
		vys = append(vys, vys[i-1].RedMul(y))
	}
	ys := NewFieldVector(vys)
	z := Hash(b128.Bytes(y.Int))

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
		parseBigInt2ABI_Bytes32(z),
		parsePoint2ABI_Bytes32_2(pcment[0]),
		parsePoint2ABI_Bytes32_2(pcment[1]),
	)

	var x = Hash(hex.EncodeToString(xbytes))
	fmt.Println("x=", x.String())

	var evalCommit = polyCommitment.Evaluate(x)
	proof.tHat = evalCommit.GetX()
	var tauX = evalCommit.GetR()
	proof.mu = alpha.RedAdd(rho.RedMul(x))
	fmt.Println("proof.mu=", proof.mu.String())

	var k_sk = b128.RandomScalar()
	var k_b = b128.RandomScalar()
	var k_tau = b128.RandomScalar()

	var A_y = burn.params.GetG().Mul(k_sk)
	var A_b = burn.params.GetG().Mul(k_b).Add(statement.CRn.Mul(zs[0]).Mul(k_sk))
	var A_t = burn.params.GetG().Mul(k_b.RedNeg()).Add(burn.params.GetH().Mul(k_tau))
	var A_u = GEpoch(statement.Epoch).Mul(k_sk)

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
		parseBigInt2ABI_Bytes32(x),
		parsePoint2ABI_Bytes32_2(A_y),
		parsePoint2ABI_Bytes32_2(A_b),
		parsePoint2ABI_Bytes32_2(A_t),
		parsePoint2ABI_Bytes32_2(A_u),
	)

	proof.c = Hash(hex.EncodeToString(cbytes))
	fmt.Println("proof.c=", proof.c.String())

	proof.s_sk = k_sk.RedAdd(proof.c.RedMul(witness.sk))
	proof.s_b = k_b.RedAdd(proof.c.RedMul(witness.bDiff.RedMul(zs[0])))
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
		parseBigInt2ABI_Bytes32(proof.c),
	)

	var o = Hash(hex.EncodeToString(obytes))
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

	fmt.Println("ipproof=", proof.ipProof.Serialize())
	return proof
}
