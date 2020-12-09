package prover

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/hpb-project/HCash-SDK/core/ebigint"
	"github.com/hpb-project/HCash-SDK/core/utils"
	"github.com/hpb-project/HCash-SDK/core/utils/bn128"
	"math"
	"math/big"
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

	ipProof *InnerProductProof
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

func (this ZetherProver) RecursivePolynomials(plist [][]*ebigint.NBigInt, accum *Polynomial,
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

func (this ZetherProver) GenerateProof(statement map[string]interface{}, witness map[string]interface{}) *ZetherProof {
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
	proof.BA = this.params.Commit(alpha, aL, aR)

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
	proof.BS = this.params.Commit(rho, sL, sR)

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
	this.RecursivePolynomials(P, NewPolynomial(nil), a.GetVector()[0:m], b.GetVector()[0:m])
	this.RecursivePolynomials(Q, NewPolynomial(nil), a.GetVector()[m:], b.GetVector()[m:])

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
			b := this.params.GetG()
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
			b := this.params.GetG()
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
			b := this.params.GetG()
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
			b := this.params.GetG()
			c := omega[k]
			proof.y_XG[k] = b128.G1.MulScalar(b, c.Int)
		}
	}
	var vPow = ebigint.ToNBigInt(big.NewInt(1)).ToRed(b128.Q())
	t_vBTransfer := witness["bTransfer"].(*ebigint.NBigInt)
	for i := 0; i < N; i++ {
		a1 := this.params.GetG()
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
				x, y := b128.Serialize(proof.CLnG[i])
				v_CLnG[i] = [2]string{x, y}
			}
			//proof.CRnG.map(bn128.serialize),
			v_CRnG := make([][2]string, len(proof.CRnG))
			for i := 0; i < len(v_CRnG); i++ {
				x, y := b128.Serialize(proof.CRnG[i])
				v_CRnG[i] = [2]string{x, y}
			}
			//proof.C_0G.map(bn128.serialize),
			v_C_0G := make([][2]string, len(proof.C_0G))
			for i := 0; i < len(v_C_0G); i++ {
				x, y := b128.Serialize(proof.C_0G[i])
				v_C_0G[i] = [2]string{x, y}
			}
			//proof.DG.map(bn128.serialize),
			v_DG := make([][2]string, len(proof.DG))
			for i := 0; i < len(v_DG); i++ {
				x, y := b128.Serialize(proof.DG[i])
				v_DG[i] = [2]string{x, y}
			}
			//proof.y_0G.map(bn128.serialize),
			v_y_0G := make([][2]string, len(proof.y_0G))
			for i := 0; i < len(v_y_0G); i++ {
				x, y := b128.Serialize(proof.y_0G[i])
				v_y_0G[i] = [2]string{x, y}
			}
			//proof.gG.map(bn128.serialize),
			v_gG := make([][2]string, len(proof.gG))
			for i := 0; i < len(v_gG); i++ {
				x, y := b128.Serialize(proof.gG[i])
				v_gG[i] = [2]string{x, y}
			}
			//proof.C_XG.map(bn128.serialize),
			v_C_XG := make([][2]string, len(proof.C_XG))
			for i := 0; i < len(v_C_XG); i++ {
				x, y := b128.Serialize(proof.C_XG[i])
				v_C_XG[i] = [2]string{x, y}
			}
			//proof.y_XG.map(bn128.serialize),
			v_y_XG := make([][2]string, len(proof.y_XG))
			for i := 0; i < len(v_y_XG); i++ {
				x, y := b128.Serialize(proof.y_XG[i])
				v_y_XG[i] = [2]string{x, y}
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
			w = utils.Hash(hex.EncodeToString(bytes))
		}
	}
	proof.f = b.Times(w).Add(a)
	{
		a1 := fq.Mul(r_B.Int, w.Int)
		a2 := fq.Add(a1, r_A.Int)
		proof.z_A = ebigint.ToNBigInt(a2).ToRed(w.GetRed())
	}
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
		y = utils.Hash(hex.EncodeToString(bytes))
	}
	var vys = make([]*ebigint.NBigInt, 0)
	{
		vys = append(vys, ebigint.NewNBigInt(1).ToRed(b128.Q()))
		for i := 1; i < 64; i++ {
			p := vys[i-1]
			nv := ebigint.ToNBigInt(fq.Mul(p.Int, y.Int)).ToRed(b128.Q())
			vys = append(vys, nv)
		}
	}
	ys := NewFieldVector(vys)
	z := utils.Hash(b128.Bytes(y.Int))
	zs := make([]*ebigint.NBigInt, 0)
	{
		zs = append(zs, ebigint.ToNBigInt(fq.Exp(z.Int, big.NewInt(2))).ToRed(z.GetRed()))
		zs = append(zs, ebigint.ToNBigInt(fq.Exp(z.Int, big.NewInt(3))).ToRed(z.GetRed()))
	}
	var twos = make([]*ebigint.NBigInt, 0)
	var v_twoTimesZs = make([]*ebigint.NBigInt, 0)
	{
		twos = append(twos, ebigint.NewNBigInt(1).ToRed(b128.Q()))
		for i := 1; i < 32; i++ {
			a1 := twos[i-1]
			a2 := big.NewInt(2)

			b := fq.Mul(a1.Int, a2)
			twos = append(twos, ebigint.ToNBigInt(b).ToRed(a1.GetRed()))
		}

		for i := 0; i < 2; i++ {
			for j := 0; j < 32; j++ {
				a1 := zs[i]
				a2 := twos[j]
				p := fq.Mul(a1.Int, a2.Int)
				v_twoTimesZs = append(v_twoTimesZs, ebigint.ToNBigInt(p).ToRed(a1.GetRed()))
			}
		}
	}
	twoTimesZs := NewFieldVector(v_twoTimesZs)
	nz := ebigint.ToNBigInt(fq.Neg(z.Int)).ToRed(z.GetRed())
	var lPoly = NewFieldVectorPolynomial(aL.Plus(nz), sL)
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
			x, y := b128.Serialize(vt[i])
			vvt = append(vvt, [2]string{x, y})
		}
		bytes, _ = arguments.Pack(
			b128.Bytes(z.Int),
			vvt[0],
			vvt[1],
		)
		x = utils.Hash(hex.EncodeToString(bytes))
	}
	var evalCommit = polyCommitment.Evaluate(x)
	proof.tHat = evalCommit.GetX()
	var tauX = evalCommit.GetR()
	{
		a1 := fq.Mul(rho.Int, x.Int)
		proof.mu = ebigint.ToNBigInt(fq.Add(alpha.Int, a1)).ToRed(alpha.GetRed())
	}
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
			{ //CRnR = CRnR.add(params.getG().mul(phi[k].redNeg().redMul(wPow)));
				a1 := fq.Neg(phi[k].Int)
				a2 := fq.Mul(a1, wPow.Int)
				a3 := b128.G1.MulScalar(this.params.GetG(), a2)
				CRnR = b128.G1.Add(CRnR, a3)
			}
			{ //DR = DR.add(params.getG().mul(chi[k].redNeg().redMul(wPow)));
				a1 := fq.Neg(chi[k].Int)
				a2 := fq.Mul(a1, wPow.Int)
				a3 := b128.G1.MulScalar(this.params.GetG(), a2)
				DR = b128.G1.Add(DR, a3)
			}
			{ //y_0R = y_0R.add(statement['y'].getVector()[witness['index'][0]].mul(psi[k].redNeg().redMul(wPow)));
				t_vy := statement["y"].(*GeneratorVector)
				t_vIndex := witness["index"].([]int)
				a1 := t_vy.GetVector()[t_vIndex[0]]
				a2 := fq.Neg(psi[k].Int)
				a3 := fq.Mul(a2, wPow.Int)
				a4 := b128.G1.MulScalar(a1, a3)
				y_0R = b128.G1.Add(y_0R, a4)
			}
			{ //gR = gR.add(params.getG().mul(psi[k].redNeg().redMul(wPow)));
				a1 := fq.Neg(psi[k].Int)
				a2 := fq.Mul(a1, wPow.Int)
				a3 := b128.G1.MulScalar(this.params.GetG(), a2)
				gR = b128.G1.Add(gR, a3)
			}
			{ //y_XR = y_XR.add(proof.y_XG[k].mul(wPow.neg()));
				a1 := proof.y_XG[k]
				a2 := wPow.Neg(wPow.Int)
				a3 := b128.G1.MulScalar(a1, a2)
				y_XR = b128.G1.Add(y_XR, a3)
			}
			p = p.Add(NP[k].Times(wPow))
			q = q.Add(NQ[k].Times(wPow))
			wPow = ebigint.ToNBigInt(fq.Mul(wPow.Int, w.Int)).ToRed(wPow.GetRed())
		}
		{ //CRnR = CRnR.add(statement['CRn'].getVector()[witness['index'][0]].mul(wPow));
			t_vCRn := statement["CRn"].(*GeneratorVector)
			t_vIndex := witness["index"].([]int)
			a1 := t_vCRn.GetVector()[t_vIndex[0]]
			a2 := b128.G1.MulScalar(a1, wPow.Int)
			CRnR = b128.G1.Add(CRnR, a2)
		}
		{ //y_0R = y_0R.add(statement['y'].getVector()[witness['index'][0]].mul(wPow));
			t_vy := statement["y"].(*GeneratorVector)
			t_vIndex := witness["index"].([]int)
			a1 := t_vy.GetVector()[t_vIndex[0]]
			a2 := b128.G1.MulScalar(a1, wPow.Int)
			y_0R = b128.G1.Add(y_0R, a2)
		}
		{ //DR = DR.add(statement['D'].mul(wPow));
			t_vD := statement["D"].(utils.Point)
			a1 := b128.G1.MulScalar(t_vD, wPow.Int)
			DR = b128.G1.Add(DR, a1)
		}
		{ //gR = gR.add(params.getG().mul(wPow));
			a1 := this.params.GetG()
			a2 := b128.G1.MulScalar(a1, wPow.Int)
			gR = b128.G1.Add(gR, a2)
		}
		{
			//p = p.add(new FieldVector(Array.from({ length: N }).map((_, i) => i == witness['index'][0] ? wPow : new BN().toRed(bn128.q))));
			vtp := make([]*ebigint.NBigInt, N)
			t_vIndex := witness["index"].([]int)
			for i := 0; i < N; i++ {
				if i == t_vIndex[0] {
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
				if i == t_vIndex[1] {
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
		t_vy := statement["y"].(*GeneratorVector)
		var convolver = NewConvolver()
		var y_p = convolver.Convolution_Point(p, t_vy)
		var y_q = convolver.Convolution_Point(q, t_vy)
		vPow = ebigint.NewNBigInt(1).ToRed(b128.Q())
		for i := 0; i < N; i++ {
			var y_poly *GeneratorVector
			if i%2 != 0 {
				y_poly = y_q
			} else {
				y_poly = y_p
			}
			idx := int(math.Floor(float64(i) / 2))
			a1 := y_poly.GetVector()[idx]
			a2 := b128.G1.MulScalar(a1, vPow.Int)
			y_XR = b128.G1.Add(y_XR, a2)
			if i > 0 {
				vPow = ebigint.ToNBigInt(fq.Mul(vPow.Int, v.Int)).ToRed(vPow.GetRed())
			}
		}
	}
	var k_sk = b128.RanddomScalar()
	var k_r = b128.RanddomScalar()
	var k_b = b128.RanddomScalar()
	var k_tau = b128.RanddomScalar()

	var A_y = b128.G1.MulScalar(gR, k_sk.Int)
	var A_D = b128.G1.MulScalar(this.params.GetG(), k_r.Int)
	var A_b utils.Point
	{ //var A_b = params.getG().mul(k_b).add(DR.mul(zs[0].redNeg()).add(CRnR.mul(zs[1])).mul(k_sk));
		a1 := b128.G1.MulScalar(this.params.GetG(), k_b.Int)

		b1 := b128.G1.MulScalar(DR, fq.Neg(zs[0].Int))
		b2 := b128.G1.MulScalar(CRnR, zs[1].Int)
		b3 := b128.G1.Add(b1, b2)

		A_b = b128.G1.Add(a1, b3)
	}
	var A_X = b128.G1.MulScalar(y_XR, k_r.Int)
	var A_t utils.Point
	{
		a1 := b128.G1.MulScalar(this.params.GetG(), fq.Neg(k_b.Int))
		a2 := b128.G1.MulScalar(this.params.GetH(), k_tau.Int)
		A_t = b128.G1.Add(a1, a2)
	}
	var A_u utils.Point
	{ //var A_u = utils.gEpoch(statement['epoch']).mul(k_sk);
		vepoch := statement["epoch"].(uint)
		A_u = b128.G1.MulScalar(utils.GEpoch(vepoch), k_sk.Int)
	}
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
		x_A_y, y_A_y := b128.Serialize(A_y)
		x_A_D, y_A_D := b128.Serialize(A_D)
		x_A_b, y_A_b := b128.Serialize(A_b)
		x_A_X, y_A_X := b128.Serialize(A_X)
		x_A_t, y_A_t := b128.Serialize(A_t)
		x_A_u, y_A_u := b128.Serialize(A_u)
		bytes, _ = arguments.Pack(
			b128.Bytes(x.Int),
			[2]string{x_A_y, y_A_y},
			[2]string{x_A_D, y_A_D},
			[2]string{x_A_b, y_A_b},
			[2]string{x_A_X, y_A_X},
			[2]string{x_A_t, y_A_t},
			[2]string{x_A_u, y_A_u},
		)
		proof.c = utils.Hash(hex.EncodeToString(bytes))
	}
	{
		t_vsk := witness["sk"].(*ebigint.NBigInt)
		t_vr := witness["r"].(*ebigint.NBigInt)
		t_vBTransfer = witness["bTransfer"].(*ebigint.NBigInt)
		t_vbDiff := witness["bDiff"].(*ebigint.NBigInt)

		red := proof.c.GetRed()
		a1 := fq.Mul(proof.c.Int, t_vsk.Int)
		proof.s_sk = ebigint.ToNBigInt(fq.Add(k_sk.Int, a1)).ToRed(red)

		a1 = fq.Mul(proof.c.Int, t_vr.Int)
		proof.s_r = ebigint.ToNBigInt(fq.Add(k_r.Int, a1)).ToRed(red)

		//proof.s_b = k_b.redAdd(proof.c.redMul(witness['bTransfer'].redMul(zs[0]).redAdd(witness['bDiff'].redMul(zs[1])).redMul(wPow)));
		b1 := fq.Mul(t_vBTransfer.Int, zs[0].Int)
		b2 := fq.Mul(t_vbDiff.Int, zs[1].Int)
		c1 := fq.Add(b1, b2)
		c2 := fq.Mul(c1, wPow.Int)
		d1 := fq.Mul(proof.c.Int, c2)
		proof.s_b = ebigint.ToNBigInt(fq.Add(k_b.Int, d1)).ToRed(red)

		//proof.s_tau = k_tau.redAdd(proof.c.redMul(tauX.redMul(wPow)));
		m1 := fq.Mul(tauX.Int, wPow.Int)
		m2 := fq.Mul(proof.c.Int, m1)
		proof.s_tau = ebigint.ToNBigInt(fq.Add(k_tau.Int, m2)).ToRed(red)
	}
	var gs = this.params.GetGS()
	var hPrimes = this.params.GetHS().Hadamard(ys.Invert())
	var hExp = ys.Times(z).Add(twoTimesZs)
	{
		//var P = proof.BA.add(proof.BS.mul(x)).add(gs.sum().mul(z.redNeg())).add(hPrimes.commit(hExp)); // rename of P
		a1 := b128.G1.MulScalar(proof.BS, x.Int)
		a2 := b128.G1.Add(proof.BA, a1)

		b1 := b128.G1.MulScalar(gs.Sum(), fq.Neg(z.Int))
		b2 := b128.G1.Add(a2, b1)

		c1 := hPrimes.Commit(hExp)
		c2 := b128.G1.Add(b2, c1)
		var t_P = c2

		//P = P.add(params.getH().mul(proof.mu.redNeg())); // Statement P of protocol 1. should this be included in the calculation of v...?
		a1 = b128.G1.MulScalar(this.params.GetH(), fq.Neg(proof.mu.Int))
		t_P = b128.G1.Add(t_P, a1)

		arguments = abi.Arguments{
			{
				Type: bytes32_T,
			},
		}
		bytes, _ = arguments.Pack(
			b128.Bytes(proof.c.Int),
		)
		o := utils.Hash(hex.EncodeToString(bytes))

		//var u_x = params.getG().mul(o); // Begin Protocol 1. this is u^x in Protocol 1. use our g for their u, our o for their x.
		var u_x = b128.G1.MulScalar(this.params.GetG(), o.Int)
		//P = P.add(u_x.mul(proof.tHat)); // corresponds to P' in protocol 1.
		t_P = b128.G1.Add(t_P, b128.G1.MulScalar(u_x, proof.tHat.Int))

		var primeBase = NewGeneratorParams(u_x, gs, hPrimes)

		var ipStatement = make(map[string]interface{})
		ipStatement["primeBase"] = primeBase
		ipStatement["P"] = t_P
		var ipWitness = make(map[string]interface{})
		ipWitness["l"] = lPoly.Evaluate(x)
		ipWitness["r"] = rPoly.Evaluate(x)

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
