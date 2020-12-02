package prover

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/hpb-project/HCash-SDK/core/ebigint"
	"github.com/hpb-project/HCash-SDK/core/utils"
	"github.com/hpb-project/HCash-SDK/core/utils/bn128"
	"math/big"
	"sort"
	"strings"
)

type ZetherProof struct {
	BA utils.Point
	BS utils.Point
	A  utils.Point
	B  utils.Point

	CLnG []utils.Point
	CRnG []utils.Point
	C_0G []utils.Point
	DG   []utils.Point
	y_0G []utils.Point
	gG   []utils.Point
	C_XG []utils.Point
	y_XG []utils.Point

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

	ipProof InnerProductProof
}

func (z ZetherProof) Serialize() string {
	b128 := utils.NewBN128()
	result := "0x"
	result += b128.Representation(z.BA)[2:]
	result += b128.Representation(z.BS)[2:]
	result += b128.Representation(z.A)[2:]
	result += b128.Representation(z.B)[2:]

	for _, CLnG_k := range z.CLnG {
		result += b128.Representation(CLnG_k)
	}

	for _, CRnG_k := range z.CRnG {
		result += b128.Representation(CRnG_k)
	}

	for _, C_0G_k := range z.C_0G {
		result += b128.Representation(C_0G_k)
	}
	for _, DG_k := range z.DG {
		result += b128.Representation(DG_k)
	}
	for _, y_0G_k := range z.y_0G {
		result += b128.Representation(y_0G_k)
	}
	for _, gG_k := range z.gG {
		result += b128.Representation(gG_k)
	}
	for _, C_XG_k := range z.C_XG {
		result += b128.Representation(C_XG_k)
	}
	for _, y_XG_k := range z.y_XG {
		result += b128.Representation(y_XG_k)
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

func (z ZetherProver) RecursivePolynomials(plist [][]*ebigint.NBigInt, accum *Polynomial,
	a []*ebigint.NBigInt, b []*ebigint.NBigInt) {
	if a == nil || len(a) == 0 {
		plist = append(plist, accum.coefficients)
		return
	}
	b128 := utils.NewBN128()
	fq := bn128.NewFq(b128.Q().Number())

	var aTop = a[len(a)-1]
	a = a[0 : len(a)-1]

	var bTop = b[len(b)-1]
	b = b[0 : len(b)-1]

	var coefficients_1 = make([]*ebigint.NBigInt, 0)
	t1 := ebigint.ToNBigInt(fq.Neg(aTop.Int)).ToRed(aTop.GetRed())
	t2 := ebigint.ToNBigInt(fq.Sub(ebigint.ToNBigInt(big.NewInt(1)).ToRed(b128.Q()).Int, bTop.Int)).ToRed(b128.Q())
	coefficients_1 = append(coefficients_1, t1)
	coefficients_1 = append(coefficients_1, t2)
	var left = NewPolynomial(coefficients_1)

	var coefficients_2 = make([]*ebigint.NBigInt, 0)
	coefficients_2 = append(coefficients_2, aTop)
	coefficients_2 = append(coefficients_2, bTop)
	var right = NewPolynomial(coefficients_2)

	z.RecursivePolynomials(plist, accum.Mul(left), a, b)
	z.RecursivePolynomials(plist, accum.Mul(right), a, b)

	a = append(a, aTop)
	b = append(b, bTop)
}

func Reverse(s []string) []string {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func (z ZetherProver) GenerateProof(statement map[string]interface{}, witness map[string]interface{}) *ZetherProof {
	proof := &ZetherProof{}

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
	vCLn := statement["CLn"].([][2]string) //{{x1,y1}, {x2,y2}...}
	vCRn := statement["CRn"].([][2]string)
	vC := statement["C"].([][2]string)
	vD := statement["D"].([2]string)
	vy := statement["y"].([][2]string)
	vepoch := statement["epoch"].(uint)

	bytes, _ := arguments.Pack(
		vCLn,
		vCRn,
		vC,
		vD,
		vy,
		vepoch)
	b128 := utils.NewBN128()
	var statementHash = utils.Hash(hex.EncodeToString(bytes))
	{
		gv := make([]utils.Point, 0)
		for _, CLn := range vCLn {
			p := b128.UnSerialize(CLn[0], CLn[1])
			gv = append(gv, p)
		}
		statement["CLn"] = NewGeneratorVector(gv)
	}
	{
		gv := make([]utils.Point, 0)
		for _, CRn := range vCRn {
			p := b128.UnSerialize(CRn[0], CRn[1])
			gv = append(gv, p)
		}
		statement["CRn"] = NewGeneratorVector(gv)
	}
	{
		gv := make([]utils.Point, 0)
		for _, C := range vC {
			p := b128.UnSerialize(C[0], C[1])
			gv = append(gv, p)
		}
		statement["C"] = NewGeneratorVector(gv)
	}
	{
		statement["D"] = b128.UnSerialize(vD[0], vD[1])
	}
	{
		gv := make([]utils.Point, 0)
		for _, y := range vy {
			p := b128.UnSerialize(y[0], y[1])
			gv = append(gv, p)
		}
		statement["y"] = NewGeneratorVector(gv)
	}

	{
		vbTransfer := witness["bTransfer"].(uint)
		witness["bTransfer"] = ebigint.ToNBigInt(big.NewInt(int64(vbTransfer))).ToRed(b128.Q())
	}
	{
		vbDiff := witness["bDiff"].(uint)
		witness["bDiff"] = ebigint.ToNBigInt(big.NewInt(int64(vbDiff))).ToRed(b128.Q())
	}
	nvBTransfer := witness["bTransfer"].(*ebigint.NBigInt)
	nvBDiff := witness["bDiff"].(*ebigint.NBigInt)

	t1 := big.NewInt(0).Lsh(nvBDiff.Int, 32)
	var number = big.NewInt(0).Add(nvBTransfer.Int, t1)
	splits := strings.Split(number.Text(2), "")
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
	proof.BA = z.params.Commit(alpha, aL, aR)

	var vsL = make([]*ebigint.NBigInt, 0)
	var vsR = make([]*ebigint.NBigInt, 0)
	for i := 0; i < 64; i++ {
		r1 := b128.RanddomScalar()
		vsL = append(vsL, r1)
		r2 := b128.RanddomScalar()
		vsR = append(vsR, r2)
	}
	var sL = NewFieldVector(vsL)
	var sR = NewFieldVector(vsR)
	var rho = b128.RanddomScalar()
	proof.BS = z.params.Commit(rho, sL, sR)

	nvy := statement["y"].(*GeneratorVector)
	var N = nvy.Length()
	//if (N & (N-1)) {
	//	throw "Size must be a power of 2!"
	//}

	var m = big.NewInt(int64(N)).BitLen() - 1
	var r_A = b128.RanddomScalar()
	var r_B = b128.RanddomScalar()

	var pa = make([]*ebigint.NBigInt, 2*m)
	for i := 0; i < 2*m; i++ {
		pa[i] = b128.RanddomScalar()
	}
	var a = NewFieldVector(pa)

	var b *FieldVector
	{
		vIndex := witness["index"].([]int)
		v1 := big.NewInt(int64(vIndex[1])).Text(2)
		v2 := big.NewInt(int64(vIndex[0])).Text(2)
		nvindex := PaddingString(v1, m) + PaddingString(v2, m)

		nsplits := strings.Split(nvindex, "")
		nreversed := Reverse(nsplits)
		nArray2 := make([]*ebigint.NBigInt, len(nreversed))
		for i, r := range nreversed {
			n, _ := big.NewInt(0).SetString(r, 2)
			nArray2[i] = ebigint.ToNBigInt(n).ToRed(b128.Q())
		}
		b = NewFieldVector(nArray2)
	}
	var c = a.Hadamard(b.Times(ebigint.ToNBigInt(big.NewInt(2)).ToRed(b128.Q())).Negate().Plus(ebigint.ToNBigInt(big.NewInt(1)).ToRed(b128.Q())))
	var d = a.Hadamard(a).Negate()
	var e, f *FieldVector
	{
		av := a.GetVector()
		evector := make([]*ebigint.NBigInt, 0)
		evector = append(evector, ebigint.ToNBigInt(fq.Mul(av[0].Int, av[m].Int)).ToRed(av[0].GetRed()))
		evector = append(evector, ebigint.ToNBigInt(fq.Mul(av[0].Int, av[m].Int)).ToRed(av[0].GetRed()))
		e = NewFieldVector(evector)

		bv := b.GetVector()
		fvector := make([]*ebigint.NBigInt, 0)
		fvector = append(fvector, av[bv[0].Int64()*int64(m)])
		pd := bv[m].Int64() * int64(m)

		fvector = append(fvector, ebigint.ToNBigInt(fq.Neg(av[pd].Int)).ToRed(av[pd].GetRed()))
		f = NewFieldVector(fvector)
	}

	proof.A = z.params.Commit(r_A, a.Concat(d).Concat(e), nil)
	proof.B = z.params.Commit(r_B, b.Concat(c).Concat(f), nil)

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
		vBAx, vBAy := b128.Serialize(proof.BA)
		vBSx, vBSy := b128.Serialize(proof.BS)
		vAx, vAy := b128.Serialize(proof.A)
		vBx, vBy := b128.Serialize(proof.B)
		bytes, _ = arguments.Pack(
			b128.Bytes(statementHash.Int),
			[]string{vBAx, vBAy},
			[]string{vBSx, vBSy},
			[]string{vAx, vAy},
			[]string{vBx, vBy},
		)
		v = utils.Hash(hex.EncodeToString(bytes))
	}
	var phi, chi, psi, omega = make([]*ebigint.NBigInt, m), make([]*ebigint.NBigInt, m), make([]*ebigint.NBigInt, m), make([]*ebigint.NBigInt, m)
	for i := 0; i < m; i++ {
		phi[i] = b128.RanddomScalar()
		chi[i] = b128.RanddomScalar()
		psi[i] = b128.RanddomScalar()
		omega[i] = b128.RanddomScalar()
	}
	var P, Q = make([][]*ebigint.NBigInt, 0), make([][]*ebigint.NBigInt, 0)
	z.RecursivePolynomials(P, NewPolynomial(nil), a.GetVector()[0:m], b.GetVector()[0:m])
	z.RecursivePolynomials(Q, NewPolynomial(nil), a.GetVector()[m:], b.GetVector()[m:])

	NP, NQ := make([]*FieldVector, m), make([]*FieldVector, m)
	{

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
		t_vCLn := statement["CLn"].(*GeneratorVector)
		t_vy := statement["y"].(*GeneratorVector)
		t_vIndex := witness["index"].([]int)
		proof.CLnG = make([]utils.Point, m)
		for k := 0; k < m; k++ {
			a1 := t_vCLn.Commit(NP[k])
			b := t_vy.GetVector()[t_vIndex[0]]
			c := phi[k]
			a2 := b128.G1.MulScalar(b, c.Int)
			proof.CLnG[k] = b128.G1.Add(a1, a2)
		}

		//proof.CRnG = Array.from({ length: m }).map((_, k) => statement['CRn'].commit(P[k]).add(params.getG().mul(phi[k])));
		t_vCRn := statement["CRn"].(*GeneratorVector)
		proof.CRnG = make([]utils.Point, m)
		for k := 0; k < m; k++ {
			a1 := t_vCRn.Commit(NP[k])
			b := z.params.GetG()
			c := phi[k]
			a2 := b128.G1.MulScalar(b, c.Int)
			proof.CRnG[k] = b128.G1.Add(a1, a2)
		}
		//proof.C_0G = Array.from({ length: m }).map((_, k) => statement['C'].commit(P[k]).add(statement['y'].getVector()[witness['index'][0]].mul(chi[k])));
		t_vC := statement["C"].(*GeneratorVector)
		proof.C_0G = make([]utils.Point, m)
		for k := 0; k < m; k++ {
			a1 := t_vC.Commit(NP[k])
			b := t_vy.GetVector()[t_vIndex[0]]
			c := chi[k]
			a2 := b128.G1.MulScalar(b, c.Int)
			proof.C_0G[k] = b128.G1.Add(a1, a2)
		}

		//proof.DG = Array.from({ length: m }).map((_, k) => params.getG().mul(chi[k]));
		proof.DG = make([]utils.Point, m)
		for k := 0; k < m; k++ {
			b := z.params.GetG()
			c := chi[k]
			proof.DG[k] = b128.G1.MulScalar(b, c.Int)
		}

		//proof.y_0G = Array.from({ length: m }).map((_, k) => statement['y'].commit(P[k]).add(statement['y'].getVector()[witness['index'][0]].mul(psi[k])));
		proof.y_0G = make([]utils.Point, m)
		for k := 0; k < m; k++ {
			a1 := t_vy.Commit(NP[k])
			b := t_vy.GetVector()[t_vIndex[0]]
			c := psi[k]
			a2 := b128.G1.MulScalar(b, c.Int)
			proof.DG[k] = b128.G1.Add(a1, a2)
		}

		//proof.gG = Array.from({ length: m }).map((_, k) => params.getG().mul(psi[k]));
		proof.gG = make([]utils.Point, m)
		for k := 0; k < m; k++ {
			b := z.params.GetG()
			c := psi[k]
			proof.gG[k] = b128.G1.MulScalar(b, c.Int)
		}

		//proof.C_XG = Array.from({ length: m }).map((_, k) => statement['D'].mul(omega[k]));
		t_vD := statement["D"].(utils.Point)
		proof.C_XG = make([]utils.Point, m)
		for k := 0; k < m; k++ {
			a := omega[k]
			proof.C_XG[k] = b128.G1.MulScalar(t_vD, a.Int)
		}

		//proof.y_XG = Array.from({ length: m }).map((_, k) => params.getG().mul(omega[k]));
		proof.y_XG = make([]utils.Point, m)
		for k := 0; k < m; k++ {
			b := z.params.GetG()
			c := omega[k]
			proof.y_XG[k] = b128.G1.MulScalar(b, c.Int)
		}
	}
	var vPow = ebigint.ToNBigInt(big.NewInt(1)).ToRed(b128.Q())
	t_vBTransfer := witness["bTransfer"].(*ebigint.NBigInt)
	for i := 0; i < N; i++ {
		a1 := z.params.GetG()
		a2 := fq.Mul(t_vBTransfer.Int, vPow.Int)

		var temp = b128.G1.MulScalar(a1, a2)
		var poly = NQ
		if i%2 == 0 {
			poly = NP
		}
		//proof.C_XG = proof.C_XG.map((C_XG_k, k) => C_XG_k.add(temp.mul(poly[k].getVector()[(witness['index'][0] + N - (i - i % 2)) % N].redNeg().redAdd(poly[k].getVector()[(witness['index'][1] + N - (i - i % 2)) % N]))));
		//proof.C_XG = proof.C_XG.map((C_XG_k, k) => C_XG_k.add(temp.mul(poly[k].getVector()[idx1].redNeg().redAdd(poly[k].getVector()[idx2]))));
		t_vIndex := witness["index"].([]int)
		n_C_XG := make([]utils.Point, len(proof.C_XG))
		for k, C_XG_k := range proof.C_XG {
			idx1 := (t_vIndex[0] + N - (i - i%2)) % N
			idx2 := (t_vIndex[1] + N - (i - i%2)) % N
			//C_XG_k.add(temp.mul(poly[k].getVector()[idx1].redNeg().redAdd(poly[k].getVector()[idx2])))
			b1 := poly[k].GetVector()[idx1]
			b2 := poly[k].GetVector()[idx2]
			c1 := fq.Neg(b1.Int)
			c2 := fq.Add(c1, b2.Int)
			d := b128.G1.MulScalar(temp, c2)
			n_C_XG[k] = b128.G1.Add(C_XG_k, d)
		}
		proof.C_XG = n_C_XG
		if i != 0 {
			vPow = ebigint.ToNBigInt(fq.Mul(vPow.Int, v.Int)).ToRed(b128.Q())
		}
	}

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
