package prover

import (
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/hpb-project/HCash-SDK/core/ebigint"
	"github.com/hpb-project/HCash-SDK/core/utils"
	"github.com/hpb-project/HCash-SDK/core/utils/bn128"
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

	ipProof InnerProductProof
}

func (z BurnProof) Serialize() string {
	b128 := utils.NewBN128()
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

func (burn BurnProver) GenerateProof(statement map[string]interface{}, witness map[string]interface{}) {
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
	vCLn := statement["CLn"].([2]string) //{{x1,y1}, {x2,y2}...}
	vCRn := statement["CRn"].([2]string)
	vy := statement["y"].([2]string)
	vepoch := statement["epoch"].(uint)
	vsender := statement["sender"].(string)

	bytes, _ := arguments.Pack(
		vCLn,
		vCRn,
		vy,
		vepoch,
		vsender)
	b128 := utils.NewBN128()
	var statementHash = utils.Hash(hex.EncodeToString(bytes))
	statement["CLn"] = b128.UnSerialize(vCLn[0], vCLn[1])
	statement["CRn"] = b128.UnSerialize(vCRn[0], vCRn[1])
	statement["y"] = b128.UnSerialize(vy[0], vy[1])
	vbDiff := witness["bDiff"].(uint)
	witness["bDiff"] = ebigint.ToNBigInt(big.NewInt(int64(vbDiff))).ToRed(b128.Q())

	splits := strings.Split(vbDiff.Text(2), "")
	println("len splits = ", len(splits), "xx ", splits)
	reversed := Reverse(splits)
	nArray := make([]*ebigint.NBigInt, len(reversed))
	for i, r := range reversed {
		n, _ := big.NewInt(0).SetString(r, 2)
		nArray[i] = ebigint.ToNBigInt(n).ToRed(b128.Q())
	}
	var aL = NewFieldVector(nArray)

	t := ebigint.ToNBigInt(big.NewInt(1)).ToRed(b128.Q())
	fq := bn128.NewFq(b128.Q().Number())
	t = ebigint.ToNBigInt(fq.Neg(t.Int)).ToRed(b128.Q())
	var aR = aL.Plus(t)
	var alpha = b128.RanddomScalar()
	proof.BA = burn.params.Commit(alpha, aL, aR)

	var vsL = make([]*ebigint.NBigInt, 0)
	var vsR = make([]*ebigint.NBigInt, 0)
	for i := 0; i < 32; i++ {
		r1 := b128.RanddomScalar()
		vsL = append(vsL, r1)
		r2 := b128.RanddomScalar()
		vsR = append(vsR, r2)
	}
	var sL = NewFieldVector(vsL)
	var sR = NewFieldVector(vsR)
	var rho = b128.RanddomScalar()
	proof.BS = burn.params.Commit(rho, sL, sR)

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
	statehash := b128.Bytes(statementHash.Int)
	vBAx, vBAy := b128.Serialize(proof.BA)
	vBSx, vBSy := b128.Serialize(proof.BS)

	ybytes, _ := argumentsy.Pack(
		statehash,
		[]string{vBAx, vBAy},
		[]string{vBSx, vBSy},
	)
	var y = utils.Hash(hex.EncodeToString(ybytes))

	var bn1 = ebigint.NewNBigInt(1).ToRed(b128.Q())
	var vys = make([]*ebigint.NBigInt, 0)
	vys = append(vys, bn1)
	for i := 1; i < 32; i++ {
		p := vys[i-1]
		nv := ebigint.ToNBigInt(fq.Mul(p.Int, y.Int)).ToRed(b128.Q())
		vys = append(vys, nv)
	}
	ys := NewFieldVector(vys)
	z := utils.Hash(b128.Bytes(y.Int))

	var b2 = ebigint.NewNBigInt(2)
	var bn2 = ebigint.NewNBigInt(2).ToRed(b128.Q())
	var vzs = make([]*ebigint.NBigInt, 0)
	var zs = fq.Exp(z,b2)
	vzs = append(vzs, zs) 
	var vtwos = make([]*ebigint.NBigInt, 0)
	vtwos = append(vtwos, bn1)
	for j := 1; j < 32; j++ {
		tmp1 := vtwos[j-1]
		tmp2 := ebigint.ToNBigInt(fq.Mul(tmp1.Int, bn2.Int)).ToRed(b128.Q())
		vtwos = append(vtwos, tmp2)
	}
	var twoTimesZs = NewFieldVector(vtwos).Times(zs[0])

	var newz = ebigint.ToNBigInt(fq.Neg(z.Int)).ToRed(b128.Q())
	var alplusz = aL.Plus(newz)
	var vlpoly = make([]*FieldVector, 0)
	vlpoly = append(vlpoly, alplusz)
	vlpoly = append(vlpoly, sL)
	var lPoly = NewFieldVectorPolynomial(vlpoly[1:])

	var arz = aR.Plus(z)
	var ysarz = ys.Hadamard(arz).Add(twoTimesZs)
	var srys = sR.Hadamard(ys)
	var vrpoly = make([]*FieldVector, 0)
	vrpoly = append(vrpoly, ysarz)
	vrpoly = append(vrpoly, srys)
	var rPloy = NewFieldVectorPolynomial(vrpoly[1:])
	var tPloyCoefficients = lPoly.InnerProduct(rPoly)
	var polyCommitment = NewPolyCommitment(burn.params, tPloyCoefficients)
	proof.tCommits = NewGeneratorVector(polyCommitment.GetCommitments())

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
	vp1x, vp1y := b128.Serialize(pcment[0])
	vp2x, vp2y := b128.Serialize(pcment[1])

	xbytes, _ := argumentsx.Pack(
		b128.Bytes(z.Int),
		[]string{vp1x, vp1y},
		[]string{vp2x, vp2y},
	)
	var x = utils.Hash(hex.EncodeToString(xbytes))

	var evalCommit = polyCommitment.Evaluate(x)
	proof.tHat = evalCommit.getX()
	var tauX = evalCommit.getR()

	rhox := fq.Mul(rho.Int,x.Int)
	proof.mu = fq.Add(alpha.Int,rhox)

	var k_sk = b128.RanddomScalar()
	var k_b = b128.RanddomScalar()
	var k_tau = b128.RanddomScalar()

	burng := burn.params.GetG()
	var A_y = b128.G1.MulScalar(burng, k_sk)

	Ab1 := b128.G1.MulScalar(burng, k_b)
	statecrn := b128.G1.MulScalar(vCRn.Int,zs[0])
	crnksk := b128.G1.MulScalar(statecrn,k_sk)
	var A_b = b128.G1.Add(Ab1,crnksk)

	kbneg :=  ebigint.ToNBigInt(fq.Neg(k_b)).ToRed(b128.Q())
	burngkb := b128.G1.MulScalar(burng, kbneg)
	var burnh = burn.params.GetH()
	burnhtau = b128.G1.MulScalar(burnh,k_tau)
	var A_t = b128.G1.Add(burngkb,burnhtau)

	stateepoch := util.GEpoch(vepoch)
	var A_u = b128.G1.MulScalar(stateepoch,k_sk)


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
	vy1x, vy1y := b128.Serialize(A_y)
	vb1x, vb1y := b128.Serialize(A_b)
	vt1x, vt1y := b128.Serialize(A_t)
	vu1x, vu1y := b128.Serialize(A_u)

	cbytes, _ := argumentsproofc.Pack(
		b128.Bytes(x.Int),
		[]string{vy1x, vy1y},
		[]string{vb1x, vb1y},
		[]string{vt1x, vt1y},
		[]string{vu1x, vu1y},
	)
	proof.c = utils.Hash(hex.EncodeToString(cbytes))

	var witnesssk = witness["sk"].(uint)
	proofcsk := fq.Mul(proof.c.Int,witnesssk)
	proof.s_sk = fq.Add(k_sk,proofcsk)

	witnessdiff := fq.Mul(vbDiff,zs[0])
	proofcdiff := fq.Mul(proof.c.Int,witnessdiff)
	proof.s_b = fq.Add(k_b,proofcdiff)
	
	proofctaux := fq.Mul(proof.c.Int,tauX)
	proof.s_tau = fq.Add(k_tau,prooftaux)

	var gs = burn.params.GetGS()
	var hs = burn.params.GetHS()
	invertys := ys.Invert()
	var hPrimes = hs.Hadamard(invertys)

	timesys := ys.Times(z.Int)
	var hExp = timesys.Add(twoTimesZs)

	proofbsx := b128.G1.MulScalar(proof.BS,x)
	zneg :=  ebigint.ToNBigInt(fq.Neg(z.Int)).ToRed(b128.Q()) 
	gssum := b128.G1.MulScalar(gs.Sum(),zneg)
	proofbax := b128.G1.Add(proof.BA,proofbsx)
	proofsum := b128.G1.Add(proofbax,gssum)
	hcommitexp := hPrimes.Commit(hExp)
	var P = b128.G1.Add(proofsum,hcommitexp)

	muneg :=  ebigint.ToNBigInt(fq.Neg(proof.mu)).ToRed(b128.Q()) 
	burnmu := b128.G1.MulScalar(burnh,muneg)
	P = b128.G1.Add(P,burnmu)

	argumento := abi.Arguments{
		{
			Type: bytes32_T,
		},
	}
	obytes, _ := argumento.Pack(
		b128.Bytes(proof.c),
	)
	var o = utils.Hash(hex.EncodeToString(obytes))
	var u_x = b128.G1.MulScalar(burng,o.Int)
	P = b128.G1.Add(P,b128.G1.MulScalar(u_x,proof.tHat))
	var primeBase = NewGeneratorParams(u_x,gs,hPrimes)
	proof.ipProof = generateProof(primeBase,P,lPoly,rPoly,[]utils.Point{}, []utils.Point{},o)
}
