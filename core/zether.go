package core

import (
	"encoding/hex"
	"errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/hpb-project/HCash-SDK/core/ebigint"
	"log"
	"math"
	"math/big"
	"strings"
)

type ZetherProof struct {
	BA Point
	BS Point
	A  Point
	B  Point

	CLnG []Point
	CRnG []Point
	C_0G []Point
	DG   []Point
	y_0G []Point
	gG   []Point
	C_XG []Point
	y_XG []Point

	f        *FieldVector
	z_A      *ebigint.NBigInt
	tCommits *GeneratorVector
	tHat     *ebigint.NBigInt
	mu       *ebigint.NBigInt

	c     *ebigint.NBigInt
	s_sk  *ebigint.NBigInt
	s_r   *ebigint.NBigInt
	s_b   *ebigint.NBigInt
	s_tau *ebigint.NBigInt

	ipProof *InnerProductProof
}

func (z ZetherProof) Serialize() string {
	result := "0x"
	result += b128.Representation(z.BA)[2:]
	result += b128.Representation(z.BS)[2:]
	result += b128.Representation(z.A)[2:]
	result += b128.Representation(z.B)[2:]

	for _, CLnG_k := range z.CLnG {
		result += b128.Representation(CLnG_k)[2:]
	}

	for _, CRnG_k := range z.CRnG {
		result += b128.Representation(CRnG_k)[2:]
	}

	for _, C_0G_k := range z.C_0G {
		result += b128.Representation(C_0G_k)[2:]
	}
	for _, DG_k := range z.DG {
		result += b128.Representation(DG_k)[2:]
	}
	for _, y_0G_k := range z.y_0G {
		result += b128.Representation(y_0G_k)[2:]
	}
	for _, gG_k := range z.gG {
		result += b128.Representation(gG_k)[2:]
	}
	for _, C_XG_k := range z.C_XG {
		result += b128.Representation(C_XG_k)[2:]
	}
	for _, y_XG_k := range z.y_XG {
		result += b128.Representation(y_XG_k)[2:]
	}

	fv := z.f.GetVector()
	for _, f_k := range fv {
		result += b128.Bytes(f_k.Int)[2:]
	}
	result += b128.Bytes(z.z_A.Int)[2:]

	tcv := z.tCommits.GetVector()
	for _, commit := range tcv {
		result += b128.Representation(commit)[2:]
	}

	result += b128.Bytes(z.tHat.Int)[2:]
	result += b128.Bytes(z.mu.Int)[2:]
	result += b128.Bytes(z.c.Int)[2:]
	result += b128.Bytes(z.s_sk.Int)[2:]
	result += b128.Bytes(z.s_r.Int)[2:]
	result += b128.Bytes(z.s_b.Int)[2:]
	result += b128.Bytes(z.s_tau.Int)[2:]

	result += z.ipProof.Serialize()[2:]

	return result
}

type ZetherProver struct {
	params   *GeneratorParams
	ipProver *InnerProductProver
}

func NewZetherProver() ZetherProver {
	params := NewGeneratorParams(int(64), nil, nil)
	return ZetherProver{
		params:   params,
		ipProver: new(InnerProductProver),
	}
}

func (this ZetherProver) RecursivePolynomials(plist [][]*ebigint.NBigInt, accum *Polynomial,
	a []*ebigint.NBigInt, b []*ebigint.NBigInt) {
	if len(a) == 0 {
		plist = append(plist, accum.coefficients)
		return
	}

	var aTop = a[len(a)-1]
	a = a[0 : len(a)-1]

	var bTop = b[len(b)-1]
	b = b[0 : len(b)-1]

	tmp_left := make([]*ebigint.NBigInt, 0)
	tmp_left = append(tmp_left, aTop.RedNeg())
	tmp_left = append(tmp_left, ebigint.NewNBigInt(1).ToRed(b128.Q()).RedSub(bTop))
	var left = NewPolynomial(tmp_left)

	tmp_right := make([]*ebigint.NBigInt, 0)
	tmp_right = append(tmp_right, aTop)
	tmp_right = append(tmp_right, bTop)
	var right = NewPolynomial(tmp_right)

	this.RecursivePolynomials(plist, accum.Mul(left), a, b)
	this.RecursivePolynomials(plist, accum.Mul(right), a, b)

	a = append(a, aTop)
	b = append(b, bTop)
}

func Reverse(s []string) []string {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

type interTransferStatement struct {
	CLn   *GeneratorVector
	CRn   *GeneratorVector
	C     *GeneratorVector
	D     Point
	Y     *GeneratorVector
	Epoch int
}

type interTransferWitness struct {
	bTransfer *ebigint.NBigInt
	bDiff     *ebigint.NBigInt
	index     []int
	sk        *ebigint.NBigInt
	r         *ebigint.NBigInt
}

func (this ZetherProver) toInnerStatement(tstatement TransferStatement) (*interTransferStatement, error) {
	statement := &interTransferStatement{}
	statement.Epoch = tstatement.Epoch

	{
		gv := make([]Point, 0)
		for _, CLn := range tstatement.CLn {
			p := b128.UnSerialize(CLn)
			gv = append(gv, p)
		}
		statement.CLn = NewGeneratorVector(gv)
	}
	{
		gv := make([]Point, 0)
		for _, CRn := range tstatement.CRn {
			p := b128.UnSerialize(CRn)
			gv = append(gv, p)
		}
		statement.CRn = NewGeneratorVector(gv)
	}
	{
		gv := make([]Point, 0)
		for _, C := range tstatement.C {
			p := b128.UnSerialize(C)
			gv = append(gv, p)
		}
		statement.C = NewGeneratorVector(gv)
	}
	{
		statement.D = b128.UnSerialize(tstatement.D)
	}
	{
		gv := make([]Point, 0)
		for _, y := range tstatement.Y {
			p := b128.UnSerialize(y)
			gv = append(gv, p)
		}
		statement.Y = NewGeneratorVector(gv)
	}

	return statement, nil
}

func (this ZetherProver) toWitness(iwitness TransferWitness) (*interTransferWitness, error) {
	witness := &interTransferWitness{}
	witness.bTransfer = ebigint.NewNBigInt(int64(iwitness.BTransfer)).ToRed(b128.Q())
	witness.bDiff = ebigint.NewNBigInt(int64(iwitness.BDiff)).ToRed(b128.Q())
	witness.index = make([]int, len(iwitness.Index))
	for i := 0; i < len(iwitness.Index); i++ {
		witness.index[i] = iwitness.Index[i]
	}

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

	str_random := iwitness.R
	if strings.HasPrefix(str_random, "0x") {
		str_random = str_random[2:]
	}
	random, ok := big.NewInt(0).SetString(str_random, 16)
	if !ok {
		return nil, errors.New("witness sk is invalid")
	} else {
		witness.sk = ebigint.ToNBigInt(random).ForceRed(b128.Q())
	}
	return witness, nil
}

func (this ZetherProver) GenerateProof(istatement TransferStatement, iwitness TransferWitness) *ZetherProof {
	var err error
	var statement *interTransferStatement
	var witness *interTransferWitness

	proof := &ZetherProof{}

	statement, err = this.toInnerStatement(istatement)
	if err != nil {
		log.Printf("to inner statement failed, err:%s\n", err.Error())
		return nil
	}

	witness, err = this.toWitness(iwitness)
	if err != nil {
		log.Printf("to inner witness failed, err:%s\n", err.Error())
		return nil
	}

	bytes32_T, _ := abi.NewType("bytes32", "", nil)
	bytes32_2ST, _ := abi.NewType("bytes32[2][]", "", nil)
	bytes32_2T, _ := abi.NewType("bytes32[2]", "", nil)
	uint256_T, _ := abi.NewType("uint256", "", nil)

	arguments := abi.Arguments{
		{
			Type: bytes32_2ST,
		},
		{
			Type: bytes32_2ST,
		},
		{
			Type: bytes32_2ST,
		},
		{
			Type: bytes32_2T,
		},
		{
			Type: bytes32_2ST,
		},
		{
			Type: uint256_T,
		},
	}
	vCLn := istatement.CLn
	vCRn := istatement.CRn
	vC := istatement.C
	vD := istatement.D
	vy := istatement.Y
	vepoch := statement.Epoch

	bytes, _ := arguments.Pack(
		vCLn,
		vCRn,
		vC,
		vD,
		vy,
		vepoch)

	var statementHash = Hash(hex.EncodeToString(bytes))

	var aL *FieldVector
	{
		t1 := big.NewInt(0).Lsh(witness.bDiff.Int, 32)
		number := big.NewInt(0).Add(witness.bTransfer.Int, t1)

		splits := strings.Split(number.Text(2), "")

		reversed := Reverse(splits)

		nArray := make([]*ebigint.NBigInt, len(reversed))
		for i, r := range reversed {
			n, _ := big.NewInt(0).SetString(r, 2)
			nArray[i] = ebigint.ToNBigInt(n).ToRed(b128.Q())
		}
		aL = NewFieldVector(nArray)
	}

	var aR = aL.Plus(ebigint.NewNBigInt(1).ToRed(b128.Q()).RedNeg())
	var alpha = b128.RanddomScalar()
	proof.BA = this.params.Commit(alpha, aL, aR)

	var vsL = make([]*ebigint.NBigInt, 64)
	var vsR = make([]*ebigint.NBigInt, 64)
	for i := 0; i < 64; i++ {
		vsL[i] = b128.RanddomScalar()
		vsR[i] = b128.RanddomScalar()
	}
	var sL = NewFieldVector(vsL)
	var sR = NewFieldVector(vsR)
	var rho = b128.RanddomScalar()
	proof.BS = this.params.Commit(rho, sL, sR)

	var N = statement.Y.Length()
	//if (N & (N-1)) {
	//	throw "Size must be a power of 2!"
	//}

	var m = big.NewInt(int64(N)).BitLen() - 1
	var r_A = b128.RanddomScalar()
	var r_B = b128.RanddomScalar()

	var a *FieldVector
	{
		var pa = make([]*ebigint.NBigInt, 2*m)
		for i := 0; i < 2*m; i++ {
			pa[i] = b128.RanddomScalar()
		}
		a = NewFieldVector(pa)
	}

	var b *FieldVector
	{
		//var b = new FieldVector((new BN(witness['index'][1]).toString(2, m) + new BN(witness['index'][0]).toString(2, m)).split("").reverse().map((i) => new BN(i, 2).toRed(bn128.q)));
		vIndex := witness.index
		v1 := big.NewInt(int64(vIndex[1])).Text(2)
		v2 := big.NewInt(int64(vIndex[0])).Text(2)
		nvindex := PaddingString(v1, m) + PaddingString(v2, m)

		nsplits := strings.Split(nvindex, "")
		nreversed := Reverse(nsplits)
		nArray := make([]*ebigint.NBigInt, len(nreversed))
		for i, r := range nreversed {
			n, _ := big.NewInt(0).SetString(r, 2)
			nArray[i] = ebigint.ToNBigInt(n).ToRed(b128.Q())
		}
		b = NewFieldVector(nArray)
	}
	var c = a.Hadamard(b.Times(ebigint.NewNBigInt(2).ToRed(b128.Q())).Negate().Plus(ebigint.NewNBigInt(1).ToRed(b128.Q())))
	var d = a.Hadamard(a).Negate()
	var e, f *FieldVector
	{
		av := a.GetVector()
		evector := make([]*ebigint.NBigInt, 0)
		evector = append(evector, av[0].RedMul(av[m]))
		evector = append(evector, av[0].RedMul(av[m]))
		e = NewFieldVector(evector)

		bv := b.GetVector()
		fvector := make([]*ebigint.NBigInt, 0)
		fvector = append(fvector, av[bv[0].Int64()*int64(m)])
		fvector = append(fvector, av[bv[m].Int64()*int64(m)].RedNeg())
		f = NewFieldVector(fvector)
	}

	proof.A = this.params.Commit(r_A, a.Concat(d).Concat(e), nil)
	proof.B = this.params.Commit(r_B, b.Concat(c).Concat(f), nil)

	var v *ebigint.NBigInt
	{
		arguments = abi.Arguments{
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

		bytes, _ = arguments.Pack(
			b128.Bytes(statementHash.Int),
			[2]string(b128.Serialize(proof.BA)),
			[2]string(b128.Serialize(proof.BS)),
			[2]string(b128.Serialize(proof.A)),
			[2]string(b128.Serialize(proof.B)),
		)
		v = Hash(hex.EncodeToString(bytes))
	}
	var phi, chi, psi, omega = make([]*ebigint.NBigInt, m), make([]*ebigint.NBigInt, m), make([]*ebigint.NBigInt, m), make([]*ebigint.NBigInt, m)
	for i := 0; i < m; i++ {
		phi[i] = b128.RanddomScalar()
		chi[i] = b128.RanddomScalar()
		psi[i] = b128.RanddomScalar()
		omega[i] = b128.RanddomScalar()
	}
	NP, NQ := make([]*FieldVector, m), make([]*FieldVector, m)
	{
		var P, Q = make([][]*ebigint.NBigInt, 0), make([][]*ebigint.NBigInt, 0)
		this.RecursivePolynomials(P, NewPolynomial(nil), a.GetVector()[0:m], b.GetVector()[0:m])
		this.RecursivePolynomials(Q, NewPolynomial(nil), a.GetVector()[m:], b.GetVector()[m:])

		for k := 0; k < m; k++ {
			tmpPv := make([]*ebigint.NBigInt, 0)
			tmpQv := make([]*ebigint.NBigInt, 0)
			for _, pi := range P {
				tmpPv = append(tmpPv, pi[k])
			}
			for _, qi := range Q {
				tmpQv = append(tmpQv, qi[k])
			}
			NP[k] = NewFieldVector(tmpPv)
			NQ[k] = NewFieldVector(tmpQv)
		}
	}

	{
		//proof.CLnG = Array.from({ length: m }).map((_, k) => statement['CLn'].commit(P[k]).add(statement['y'].getVector()[witness['index'][0]].mul(phi[k])));
		proof.CLnG = make([]Point, m)
		for k := 0; k < m; k++ {
			proof.CLnG[k] = statement.CLn.Commit(NP[k]).Add(statement.Y.GetVector()[witness.index[0]].Mul(phi[k]))
		}

		//proof.CRnG = Array.from({ length: m }).map((_, k) => statement['CRn'].commit(P[k]).add(params.getG().mul(phi[k])));
		proof.CRnG = make([]Point, m)
		for k := 0; k < m; k++ {
			proof.CRnG[k] = statement.CRn.Commit(NP[k]).Add(this.params.GetG().Mul(phi[k]))
		}

		//proof.C_0G = Array.from({ length: m }).map((_, k) => statement['C'].commit(P[k]).add(statement['y'].getVector()[witness['index'][0]].mul(chi[k])));
		proof.C_0G = make([]Point, m)
		for k := 0; k < m; k++ {
			proof.C_0G[k] = statement.C.Commit(NP[k]).Add(statement.Y.GetVector()[witness.index[0]].Mul(chi[k]))
		}

		//proof.DG = Array.from({ length: m }).map((_, k) => params.getG().mul(chi[k]));
		proof.DG = make([]Point, m)
		for k := 0; k < m; k++ {
			proof.DG[k] = this.params.GetG().Mul(chi[k])
		}

		//proof.y_0G = Array.from({ length: m }).map((_, k) => statement['y'].commit(P[k]).add(statement['y'].getVector()[witness['index'][0]].mul(psi[k])));
		proof.y_0G = make([]Point, m)
		for k := 0; k < m; k++ {
			proof.y_0G[k] = statement.Y.Commit(NP[k]).Add(statement.Y.GetVector()[witness.index[0]].Mul(psi[k]))
		}

		//proof.gG = Array.from({ length: m }).map((_, k) => params.getG().mul(psi[k]));
		proof.gG = make([]Point, m)
		for k := 0; k < m; k++ {
			proof.gG[k] = this.params.GetG().Mul(psi[k])
		}

		//proof.C_XG = Array.from({ length: m }).map((_, k) => statement['D'].mul(omega[k]));
		proof.C_XG = make([]Point, m)
		for k := 0; k < m; k++ {
			proof.C_XG[k] = statement.D.Mul(omega[k])
		}

		//proof.y_XG = Array.from({ length: m }).map((_, k) => params.getG().mul(omega[k]));
		proof.y_XG = make([]Point, m)
		for k := 0; k < m; k++ {
			proof.y_XG[k] = this.params.GetG().Mul(omega[k])
		}
	}
	var vPow = ebigint.NewNBigInt(1).ToRed(b128.Q())
	for i := 0; i < N; i++ {
		var temp = this.params.GetG().Mul(witness.bTransfer.RedMul(vPow))
		var poly = NQ
		if i%2 == 0 {
			poly = NP
		}
		//proof.C_XG = proof.C_XG.map((C_XG_k, k) => C_XG_k.add(temp.mul(poly[k].getVector()[(witness['index'][0] + N - (i - i % 2)) % N].redNeg().redAdd(poly[k].getVector()[(witness['index'][1] + N - (i - i % 2)) % N]))));
		n_C_XG := make([]Point, len(proof.C_XG))
		for k, C_XG_k := range proof.C_XG {
			n_C_XG[k] = C_XG_k.Add(temp.Mul(poly[k].GetVector()[(witness.index[0]+N-(i-i%2))%N].RedNeg().RedAdd(poly[k].GetVector()[(witness.index[1]+N-(i-i%2))%N])))
		}

		proof.C_XG = n_C_XG
		if i != 0 {
			vPow = vPow.RedMul(v)
		}
	}
	var w *ebigint.NBigInt
	{
		{
			arguments = abi.Arguments{
				{
					Type: bytes32_T,
				},
				{
					Type: bytes32_2ST,
				},
				{
					Type: bytes32_2ST,
				},
				{
					Type: bytes32_2ST,
				},
				{
					Type: bytes32_2ST,
				},
				{
					Type: bytes32_2ST,
				},
				{
					Type: bytes32_2ST,
				},
				{
					Type: bytes32_2ST,
				},
				{
					Type: bytes32_2ST,
				},
			}
			//proof.CLnG.map(bn128.serialize),
			v_CLnG := make([][2]string, len(proof.CLnG))
			for i := 0; i < len(v_CLnG); i++ {
				v_CLnG[i] = [2]string(b128.Serialize(proof.CLnG[i]))
			}
			//proof.CRnG.map(bn128.serialize),
			v_CRnG := make([][2]string, len(proof.CRnG))
			for i := 0; i < len(v_CRnG); i++ {
				v_CRnG[i] = [2]string(b128.Serialize(proof.CRnG[i]))
			}
			//proof.C_0G.map(bn128.serialize),
			v_C_0G := make([][2]string, len(proof.C_0G))
			for i := 0; i < len(v_C_0G); i++ {
				v_C_0G[i] = [2]string(b128.Serialize(proof.C_0G[i]))
			}
			//proof.DG.map(bn128.serialize),
			v_DG := make([][2]string, len(proof.DG))
			for i := 0; i < len(v_DG); i++ {
				v_DG[i] = [2]string(b128.Serialize(proof.DG[i]))
			}
			//proof.y_0G.map(bn128.serialize),
			v_y_0G := make([][2]string, len(proof.y_0G))
			for i := 0; i < len(v_y_0G); i++ {
				v_y_0G[i] = [2]string(b128.Serialize(proof.y_0G[i]))
			}
			//proof.gG.map(bn128.serialize),
			v_gG := make([][2]string, len(proof.gG))
			for i := 0; i < len(v_gG); i++ {
				v_gG[i] = [2]string(b128.Serialize(proof.gG[i]))
			}
			//proof.C_XG.map(bn128.serialize),
			v_C_XG := make([][2]string, len(proof.C_XG))
			for i := 0; i < len(v_C_XG); i++ {
				v_C_XG[i] = [2]string(b128.Serialize(proof.C_XG[i]))
			}
			//proof.y_XG.map(bn128.serialize),
			v_y_XG := make([][2]string, len(proof.y_XG))
			for i := 0; i < len(v_y_XG); i++ {
				v_y_XG[i] = [2]string(b128.Serialize(proof.y_XG[i]))
			}

			bytes, _ = arguments.Pack(
				b128.Bytes(v.Int),
				v_CLnG,
				v_CRnG,
				v_C_0G,
				v_DG,
				v_y_0G,
				v_gG,
				v_C_XG,
				v_y_XG,
			)
			w = Hash(hex.EncodeToString(bytes))
		}
	}
	proof.f = b.Times(w).Add(a)
	proof.z_A = r_B.RedMul(w).RedAdd(r_A)

	var y *ebigint.NBigInt
	{
		arguments = abi.Arguments{
			{
				Type: bytes32_T,
			},
		}

		bytes, _ = arguments.Pack(
			b128.Bytes(w.Int),
		)
		y = Hash(hex.EncodeToString(bytes))
	}
	var vys = make([]*ebigint.NBigInt, 0)
	{
		vys = append(vys, ebigint.NewNBigInt(1).ToRed(b128.Q()))
		for i := 1; i < 64; i++ {
			vys = append(vys, vys[i-1].RedMul(y))
		}
	}
	ys := NewFieldVector(vys)
	z := Hash(b128.Bytes(y.Int))
	zs := make([]*ebigint.NBigInt, 0)
	{
		zs = append(zs, z.RedExp(big.NewInt(2)))
		zs = append(zs, z.RedExp(big.NewInt(3)))
	}
	var twos = make([]*ebigint.NBigInt, 0)
	var v_twoTimesZs = make([]*ebigint.NBigInt, 0)
	{
		twos = append(twos, ebigint.NewNBigInt(1).ToRed(b128.Q()))
		for i := 1; i < 32; i++ {
			twos = append(twos, twos[i-1].RedMul(ebigint.NewNBigInt(2).ToRed(b128.Q())))
		}

		for i := 0; i < 2; i++ {
			for j := 0; j < 32; j++ {
				v_twoTimesZs = append(v_twoTimesZs, zs[i].RedMul(twos[j]))
			}
		}
	}
	twoTimesZs := NewFieldVector(v_twoTimesZs)

	var lPoly = NewFieldVectorPolynomial(aL.Plus(z.RedNeg()), sL)
	var rPoly = NewFieldVectorPolynomial(ys.Hadamard(aR.Plus(z)).Add(twoTimesZs), sR.Hadamard(ys))
	var tPolyCoefficients = lPoly.InnerProduct(rPoly)
	var polyCommitment = NewPolyCommitment(*this.params, tPolyCoefficients)

	proof.tCommits = NewGeneratorVector(polyCommitment.GetCommitments())

	var x *ebigint.NBigInt
	{
		arguments = abi.Arguments{
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
		vt := polyCommitment.GetCommitments()
		vvt := make([][2]string, 0)
		for i := 0; i < len(vt); i++ {
			vvt = append(vvt, [2]string(b128.Serialize(vt[i])))
		}
		bytes, _ = arguments.Pack(
			b128.Bytes(z.Int),
			vvt[0],
			vvt[1],
		)
		x = Hash(hex.EncodeToString(bytes))
	}
	var evalCommit = polyCommitment.Evaluate(x)
	proof.tHat = evalCommit.GetX()

	var tauX = evalCommit.GetR()
	proof.mu = alpha.RedAdd(rho.RedMul(x))

	var CRnR = b128.Zero()
	var y_0R = b128.Zero()
	var y_XR = b128.Zero()
	var DR = b128.Zero()
	var gR = b128.Zero()
	var p, q *FieldVector
	{
		v_p := make([]*ebigint.NBigInt, N)
		v_q := make([]*ebigint.NBigInt, N)
		for i := 0; i < N; i++ {
			v_p[i] = ebigint.NewNBigInt(0).ToRed(b128.Q())
			v_q[i] = ebigint.NewNBigInt(0).ToRed(b128.Q())
		}
		p = NewFieldVector(v_p)
		q = NewFieldVector(v_q)
	}
	var wPow = ebigint.NewNBigInt(1).ToRed(b128.Q())
	{
		for k := 0; k < m; k++ {
			CRnR = CRnR.Add(this.params.GetG().Mul(phi[k].RedNeg().RedMul(wPow)))
			DR = DR.Add(this.params.GetG().Mul(chi[k].RedNeg().RedMul(wPow)))
			y_0R = y_0R.Add(statement.Y.GetVector()[witness.index[0]].Mul(psi[k].RedNeg().RedMul(wPow)))
			gR = gR.Add(this.params.GetG().Mul(psi[k].RedNeg().RedMul(wPow)))
			y_XR = y_XR.Add(proof.y_XG[k].Mul(ebigint.ToNBigInt(big.NewInt(0).Neg(wPow.Int)).ToRed(wPow.GetRed())))

			p = p.Add(NP[k].Times(wPow))
			q = q.Add(NQ[k].Times(wPow))
			wPow = wPow.RedMul(w)
		}

		CRnR = CRnR.Add(statement.CRn.GetVector()[witness.index[0]].Mul(wPow))
		y_0R = y_0R.Add(statement.Y.GetVector()[witness.index[0]].Mul(wPow))
		DR = DR.Add(statement.D.Mul(wPow))
		gR = gR.Add(this.params.GetG().Mul(wPow))
		{
			//p = p.add(new FieldVector(Array.from({ length: N }).map((_, i) => i == witness['index'][0] ? wPow : new BN().toRed(bn128.q))));
			vtp := make([]*ebigint.NBigInt, N)
			for i := 0; i < N; i++ {
				if i == witness.index[0] {
					vtp[i] = wPow
				} else {
					vtp[i] = ebigint.NewNBigInt(0).ToRed(b128.Q())
				}
			}
			tp := NewFieldVector(vtp)
			p = p.Add(tp)
			//q = q.add(new FieldVector(Array.from({ length: N }).map((_, i) => i == witness['index'][1] ? wPow : new BN().toRed(bn128.q))));
			vtq := make([]*ebigint.NBigInt, N)
			for i := 0; i < N; i++ {
				if i == witness.index[1] {
					vtq[i] = wPow
				} else {
					vtq[i] = ebigint.NewNBigInt(0).ToRed(b128.Q())
				}
			}
			tq := NewFieldVector(vtq)
			q = q.Add(tq)
		}
	}
	{
		var convolver = NewConvolver()
		var y_p = convolver.Convolution_Point(p, statement.Y)
		var y_q = convolver.Convolution_Point(q, statement.Y)
		vPow = ebigint.NewNBigInt(1).ToRed(b128.Q())
		for i := 0; i < N; i++ {
			var y_poly *GeneratorVector
			if i%2 != 0 {
				y_poly = y_q
			} else {
				y_poly = y_p
			}
			idx := int(math.Floor(float64(i) / 2))
			y_XR = y_XR.Add(y_poly.GetVector()[idx].Mul(vPow))
			if i > 0 {
				vPow = vPow.RedMul(v)
			}
		}
	}
	var k_sk = b128.RanddomScalar()
	var k_r = b128.RanddomScalar()
	var k_b = b128.RanddomScalar()
	var k_tau = b128.RanddomScalar()

	var A_y = gR.Mul(k_sk)
	var A_D = this.params.GetG().Mul(k_r)
	var A_b = this.params.GetG().Mul(k_b).Add(DR.Mul(zs[0].RedNeg()).Add(CRnR.Mul(zs[1])).Mul(k_sk))
	var A_X = y_XR.Mul(k_r)
	var A_t = this.params.GetG().Mul(k_b.RedNeg()).Add(this.params.GetH().Mul(k_tau))
	var A_u = GEpoch(statement.Epoch).Mul(k_sk)

	{
		arguments = abi.Arguments{
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
			{
				Type: bytes32_2T,
			},
			{
				Type: bytes32_2T,
			},
		}
		bytes, _ = arguments.Pack(
			b128.Bytes(x.Int),
			[2]string(b128.Serialize(A_y)),
			[2]string(b128.Serialize(A_D)),
			[2]string(b128.Serialize(A_b)),
			[2]string(b128.Serialize(A_X)),
			[2]string(b128.Serialize(A_t)),
			[2]string(b128.Serialize(A_u)),
		)
		proof.c = Hash(hex.EncodeToString(bytes))
	}

	proof.s_sk = k_sk.RedAdd(proof.c.RedMul(witness.sk))
	proof.s_r = k_r.RedAdd(proof.c.RedMul(witness.r))

	proof.s_b = k_b.RedAdd(proof.c.RedMul(witness.bTransfer.RedMul(zs[0]).RedAdd(witness.bDiff.RedMul(zs[1])).RedMul(wPow)))
	proof.s_tau = k_tau.RedAdd(proof.c.RedMul(tauX.RedMul(wPow)))

	var gs = this.params.GetGS()
	var hPrimes = this.params.GetHS().Hadamard(ys.Invert())
	var hExp = ys.Times(z).Add(twoTimesZs)
	{
		var P = proof.BA.Add(proof.BS.Mul(x)).Add(gs.Sum().Mul(z.RedNeg())).Add(hPrimes.Commit(hExp))
		P = P.Add(this.params.GetH().Mul(proof.mu.RedNeg()))

		arguments = abi.Arguments{
			{
				Type: bytes32_T,
			},
		}
		bytes, _ = arguments.Pack(
			b128.Bytes(proof.c.Int),
		)
		o := Hash(hex.EncodeToString(bytes))

		var u_x = this.params.GetG().Mul(o)
		P = P.Add(u_x.Mul(proof.tHat))

		var primeBase = NewGeneratorParams(u_x, gs, hPrimes)

		var ipStatement = InnerProduct_statement{}
		ipStatement.PrimeBase = primeBase
		ipStatement.P = P
		var ipWitness = InnerProduct_witness{}
		ipWitness.L = lPoly.Evaluate(x)
		ipWitness.R = rPoly.Evaluate(x)

		proof.ipProof = this.ipProver.GenerateProof(ipStatement, ipWitness, o)
	}
	return proof
}

func PaddingString(in string, padding int) string {
	var out = in
	for {
		if len(out)%padding != 0 {
			out = "0" + out
		} else {
			break
		}
	}
	return out
}
