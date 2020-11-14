package prover

import (
	"encoding/hex"
	//"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/hpb-project/HCash-SDK/core/ebigint"
	"github.com/hpb-project/HCash-SDK/core/utils"
	"github.com/hpb-project/HCash-SDK/core/utils/bn128"
	solsha3 "github.com/miguelmota/go-solidity-sha3"
	"math/big"
)

type GeneratorParams struct {
	g  utils.Point
	h  utils.Point
	gs *GeneratorVector
	hs *GeneratorVector
}

func NewGeneratorParams(h int) *GeneratorParams {
	gp := &GeneratorParams{}

	gp.g = utils.MapInto(hex.EncodeToString(solsha3.SoliditySHA3(solsha3.String("G"))))
	gp.h = utils.MapInto(hex.EncodeToString(solsha3.SoliditySHA3(solsha3.String("H"))))

	gsInnards := make([]utils.Point, 0)
	hsInnards := make([]utils.Point, 0)
	for i := 0; i < h; i++ {
		p1 := utils.MapInto(hex.EncodeToString(solsha3.SoliditySHA3(
			solsha3.String("G"), solsha3.Uint32(i))))
		gsInnards = append(gsInnards, p1)
		p2 := utils.MapInto(hex.EncodeToString(solsha3.SoliditySHA3(
			solsha3.String("H"), solsha3.Uint32(i))))
		hsInnards = append(hsInnards, p2)
	}
	gp.gs = NewGeneratorVector(gsInnards)
	gp.hs = NewGeneratorVector(hsInnards)

	return gp
}

func (g GeneratorParams) GetG() utils.Point {
	return g.g
}

func (g GeneratorParams) GetH() utils.Point {
	return g.h
}

func (g GeneratorParams) GetGS() *GeneratorVector {
	return g.gs
}

func (g GeneratorParams) GetHS() *GeneratorVector {
	return g.hs
}

func (g *GeneratorParams) Commit(blinding *ebigint.NBigInt, gExp, hExp *FieldVector) utils.Point {
	b128 := utils.NewBN128()
	var result = b128.G1.MulScalar(g.h, blinding.Int)
	var gsVector = g.gs.GetVector()
	gexpVector := gExp.GetVector()
	for i, gexp := range gexpVector {
		t := b128.G1.MulScalar(gsVector[i], gexp.Int)
		result = b128.G1.Add(result, t)
	}

	if hExp != nil {
		var hsVector = g.hs.GetVector()
		hexpVector := hExp.GetVector()
		for i, hexp := range hexpVector {
			t := b128.G1.MulScalar(hsVector[i], hexp.Int)
			result = b128.G1.Add(result, t)
		}
	}
	return result
}

type FieldVector struct {
	vector []*ebigint.NBigInt
}

func NewFieldVector(vector []*ebigint.NBigInt) *FieldVector {
	fv := &FieldVector{}
	fv.vector = vector

	return fv
}

func (f *FieldVector) GetVector() []*ebigint.NBigInt {
	return f.vector
}

func (f *FieldVector) Length() int {
	return len(f.vector)
}

func (f *FieldVector) Slice(begin, end int) *FieldVector {
	var innards = f.vector[begin:end]
	return NewFieldVector(innards)
}

func (f *FieldVector) Add(other *FieldVector) *FieldVector {
	fq := bn128.NewFq(f.vector[0].GetRed().Number())
	var innards = other.GetVector()
	var nInnards = make([]*ebigint.NBigInt, len(f.vector))

	for i, accum := range f.vector {
		nInnards[i] = ebigint.ToNBigInt(fq.Add(accum.Int, innards[i].Int))
	}

	return NewFieldVector(nInnards)
}

func (f *FieldVector) Plus(constant *ebigint.NBigInt) *FieldVector {
	fq := bn128.NewFq(f.vector[0].GetRed().Number())
	var nInnards = make([]*ebigint.NBigInt, len(f.vector))
	for i, accum := range f.vector {
		nInnards[i] = ebigint.ToNBigInt(fq.Add(accum.Int, constant.Int))
	}

	return NewFieldVector(nInnards)
}

func (f *FieldVector) Sum() *ebigint.NBigInt {
	b128 := utils.NewBN128()
	fq := bn128.NewFq(b128.Q().Number())
	var nVectors = make([]*ebigint.NBigInt, 0)
	nVectors = append(nVectors, ebigint.ToNBigInt(big.NewInt(0)).ToRed(b128.Q()))
	for _, c := range f.vector {
		nVectors = append(nVectors, c)
	}

	var current = ebigint.ToNBigInt(big.NewInt(0)).ToRed(b128.Q())

	for _, accum := range nVectors {
		current = ebigint.ToNBigInt(fq.Add(accum.Int, current.Int))
	}

	return current
}

func (f *FieldVector) Negate() *FieldVector {
	fq := bn128.NewFq(f.vector[0].GetRed().Number())
	var nInnards = make([]*ebigint.NBigInt, len(f.vector))
	for i, accum := range f.vector {
		nInnards[i] = ebigint.ToNBigInt(fq.Neg(accum.Int))
	}

	return NewFieldVector(nInnards)
}

func (f *FieldVector) Subtract(other *FieldVector) *FieldVector {
	return f.Add(other.Negate())
}

func (f *FieldVector) Hadamard(other *FieldVector) *FieldVector {
	fq := bn128.NewFq(f.vector[0].GetRed().Number())
	var innards = other.vector
	var nInnards = make([]*ebigint.NBigInt, len(f.vector))
	for i, accum := range f.vector {
		nInnards[i] = ebigint.ToNBigInt(fq.Mul(accum.Int, innards[i].Int))
	}

	return NewFieldVector(nInnards)
}

func (f *FieldVector) Invert() *FieldVector {
	fq := bn128.NewFq(f.vector[0].GetRed().Number())
	var nInnards = make([]*ebigint.NBigInt, len(f.vector))
	for i, accum := range f.vector {
		nInnards[i] = ebigint.ToNBigInt(fq.Inverse(accum.Int))
	}

	return NewFieldVector(nInnards)
}

func (f *FieldVector) Extract(parity int) *FieldVector {
	var nInnards = make([]*ebigint.NBigInt, 0)
	for i, accum := range f.vector {
		if i%2 == parity {
			nInnards = append(nInnards, accum)
		}
	}
	return NewFieldVector(nInnards)
}

func (f *FieldVector) Flip() *FieldVector {
	var size = len(f.vector)
	var nInnards = make([]*ebigint.NBigInt, size)
	for i, _ := range f.vector {
		nInnards[i] = f.vector[(size-i)%size]
	}

	return NewFieldVector(nInnards)
}

func (f *FieldVector) Concat(other *FieldVector) *FieldVector {
	var nInnards = make([]*ebigint.NBigInt, 0)
	for _, elem := range f.vector {
		nInnards = append(nInnards, elem)
	}

	for _, elem := range other.vector {
		nInnards = append(nInnards, elem)
	}

	return NewFieldVector(nInnards)
}

func (f *FieldVector) Times(constant *ebigint.NBigInt) *FieldVector {
	fq := bn128.NewFq(f.vector[0].GetRed().Number())
	var nInnards = make([]*ebigint.NBigInt, len(f.vector))
	for i, accum := range f.vector {
		nInnards[i] = ebigint.ToNBigInt(fq.Mul(accum.Int, constant.Int))
	}

	return NewFieldVector(nInnards)
}

func (f *FieldVector) InnerProduct(other *FieldVector) *ebigint.NBigInt {
	var innards = other.GetVector()

	b128 := utils.NewBN128()
	fq := bn128.NewFq(b128.Q().Number())
	var nVectors = make([]*ebigint.NBigInt, 0)
	nVectors = append(nVectors, ebigint.ToNBigInt(big.NewInt(0)).ToRed(b128.Q()))
	for _, c := range f.vector {
		nVectors = append(nVectors, c)
	}

	var current = ebigint.ToNBigInt(big.NewInt(0)).ToRed(b128.Q())

	for i, accum := range nVectors {
		t := fq.Mul(current.Int, innards[i].Int)
		current = ebigint.ToNBigInt(fq.Add(accum.Int, t))
	}

	return current

}

type GeneratorVector struct {
	vector []utils.Point
}

func NewGeneratorVector(Innards []utils.Point) *GeneratorVector {
	gv := &GeneratorVector{}
	gv.vector = Innards
	return gv
}

func (g *GeneratorVector) GetVector() []utils.Point {
	return g.vector
}

func (g *GeneratorVector) Length() int {
	return len(g.vector)
}

func (g *GeneratorVector) Slice(begin, end int) *GeneratorVector {
	return NewGeneratorVector(g.vector[begin:end])
}

func (g *GeneratorVector) Commit(exponents *FieldVector) utils.Point {
	b128 := utils.NewBN128()
	var nVectors = make([]utils.Point, 0)
	nVectors = append(nVectors, b128.Zero())
	for _, c := range g.vector {
		nVectors = append(nVectors, c)
	}

	var innards = exponents.GetVector()
	var current = b128.Zero()

	for i, accum := range nVectors {
		t := b128.G1.MulScalar(current, innards[i].Int)
		current = b128.G1.Add(accum, t)
	}

	return current
}

func (g *GeneratorVector) Sum() utils.Point {
	b128 := utils.NewBN128()
	var nVectors = make([]utils.Point, 0)
	nVectors = append(nVectors, b128.Zero())
	for _, c := range g.vector {
		nVectors = append(nVectors, c)
	}

	var current = b128.Zero()

	for _, accum := range nVectors {
		current = b128.G1.Add(accum, current)
	}

	return current
}

func (g *GeneratorVector) Add(other *GeneratorVector) *GeneratorVector {
	b128 := utils.NewBN128()
	var innards = other.GetVector()
	var nInnards = make([]utils.Point, len(g.vector))
	for i, accum := range g.vector {
		nInnards[i] = b128.G1.Add(accum, innards[i])
	}

	return NewGeneratorVector(nInnards)
}

func (g *GeneratorVector) Hadamard(exponents *FieldVector) *GeneratorVector {
	b128 := utils.NewBN128()
	var innards = exponents.GetVector()
	var nInnards = make([]utils.Point, len(g.vector))
	for i, elem := range g.vector {
		nInnards[i] = b128.G1.MulScalar(elem, innards[i].Int)
	}

	return NewGeneratorVector(nInnards)
}

func (g *GeneratorVector) Negate() *GeneratorVector {
	b128 := utils.NewBN128()
	var nInnards = make([]utils.Point, len(g.vector))
	for i, elem := range g.vector {
		nInnards[i] = b128.G1.Neg(elem)
	}

	return NewGeneratorVector(nInnards)
}

func (g *GeneratorVector) Times(constant *ebigint.NBigInt) *GeneratorVector {
	b128 := utils.NewBN128()
	var nInnards = make([]utils.Point, len(g.vector))
	for i, elem := range g.vector {
		nInnards[i] = b128.G1.MulScalar(elem, constant.Int)
	}

	return NewGeneratorVector(nInnards)
}

func (g *GeneratorVector) Extract(parity int) *GeneratorVector {
	var nInnards = make([]utils.Point, len(g.vector))
	for i, elem := range g.vector {
		if i%2 == parity {
			nInnards = append(nInnards, elem)
		}
	}

	return NewGeneratorVector(nInnards)
}

func (g *GeneratorVector) Concat(other *GeneratorVector) *GeneratorVector {
	var nInnards = make([]utils.Point, 0)
	for _, elem := range g.vector {
		nInnards = append(nInnards, elem)
	}

	for _, elem := range other.vector {
		nInnards = append(nInnards, elem)
	}

	return NewGeneratorVector(nInnards)
}

type Convolver struct {
	unity *ebigint.NBigInt
}

func NewConvolver() *Convolver {
	c := &Convolver{}

	b128 := utils.NewBN128()
	unity, _ := big.NewInt(0).SetString("14a3074b02521e3b1ed9852e5028452693e87be4e910500c7ba9bbddb2f46edd", 16)
	c.unity = ebigint.ToNBigInt(unity).ToRed(b128.Q())

	return c
}

//
//type Something interface {
//	Length() int
//	Extract(int) Something
//}

func (c *Convolver) FFT_Scalar(input *FieldVector, inverse bool) *FieldVector {
	var length = input.Length()
	if length == 1 {
		return input
	}
	if length%2 != 0 {
		panic("Input size must be a power of 2!")
	}
	b128 := utils.NewBN128()
	fq := bn128.NewFq(c.unity.GetRed().Number())
	exp := big.NewInt(1).Lsh(big.NewInt(1), 28)
	var omega = fq.Exp(c.unity.Int, exp.Div(exp, big.NewInt(int64(length))))
	if inverse {
		omega = fq.Inverse(omega)
	}
	var even = c.FFT_Scalar(input.Extract(0), inverse)
	var odd = c.FFT_Scalar(input.Extract(1), inverse)

	var omegas = make([]*ebigint.NBigInt, 0)

	omegas = append(omegas, ebigint.ToNBigInt(big.NewInt(1)).ToRed(c.unity.GetRed()))

	for i := 1; i < length/2; i++ {
		t := fq.Mul(omegas[i-1].Int, omega)
		omegas = append(omegas, ebigint.ToNBigInt(t).ToRed(c.unity.GetRed()))
	}

	var nomegas = NewFieldVector(omegas)
	var result = even.Add(odd.Hadamard(nomegas)).Concat(even.Add(odd.Hadamard(nomegas).Negate()))
	if inverse {
		t := ebigint.ToNBigInt(big.NewInt(2)).ToRed(b128.Q())
		t = ebigint.ToNBigInt(fq.Inverse(t.Int)).ToRed(b128.Q())
		result = result.Times(t)
	}
	return result
}

func (c *Convolver) Convolution_Scalar(exponent *FieldVector, base *FieldVector) *FieldVector {
	b128 := utils.NewBN128()
	fq := bn128.NewFq(b128.Q().Number())
	size := base.Length()
	temp := c.FFT_Scalar(base, false).Hadamard(c.FFT_Scalar(exponent.Flip(), false))

	t := ebigint.ToNBigInt(big.NewInt(2)).ToRed(b128.Q())
	t = ebigint.ToNBigInt(fq.Inverse(t.Int)).ToRed(b128.Q())
	return c.FFT_Scalar(temp.Slice(0, size/2).Add(temp.Slice(size/2, size)).Times(t), true)
}

func (c *Convolver) FFT_Point(input *GeneratorVector, inverse bool) *GeneratorVector {
	var length = input.Length()
	if length == 1 {
		return input
	}
	if length%2 != 0 {
		panic("Input size must be a power of 2!")
	}
	b128 := utils.NewBN128()
	fq := bn128.NewFq(c.unity.GetRed().Number())
	exp := big.NewInt(1).Lsh(big.NewInt(1), 28)
	var omega = fq.Exp(c.unity.Int, exp.Div(exp, big.NewInt(int64(length))))
	if inverse {
		omega = fq.Inverse(omega)
	}
	var even = c.FFT_Point(input.Extract(0), inverse)
	var odd = c.FFT_Point(input.Extract(1), inverse)

	var omegas = make([]*ebigint.NBigInt, 0)

	omegas = append(omegas, ebigint.ToNBigInt(big.NewInt(1)).ToRed(c.unity.GetRed()))

	for i := 1; i < length/2; i++ {
		t := fq.Mul(omegas[i-1].Int, omega)
		omegas = append(omegas, ebigint.ToNBigInt(t).ToRed(c.unity.GetRed()))
	}

	var nomegas = NewFieldVector(omegas)
	var result = even.Add(odd.Hadamard(nomegas)).Concat(even.Add(odd.Hadamard(nomegas).Negate()))
	if inverse {
		t := ebigint.ToNBigInt(big.NewInt(2)).ToRed(b128.Q())
		t = ebigint.ToNBigInt(fq.Inverse(t.Int)).ToRed(b128.Q())
		result = result.Times(t)
	}
	return result
}

type FieldVectorPolynomial struct {
	//coefficients []*PedersenCommitment
	coefficients []*FieldVector
}

func NewFieldVectorPolynomial(coefficients []*FieldVector) *FieldVectorPolynomial {
	fvp := &FieldVectorPolynomial{
		coefficients: coefficients,
	}
	return fvp
}

func (f *FieldVectorPolynomial) GetCoefficients() []*FieldVector {
	return f.coefficients
}

func (f *FieldVectorPolynomial) Evaluate(x *ebigint.NBigInt) *FieldVector {
	result := f.coefficients[0]
	var accumulator = x
	fq := bn128.NewFq(x.GetRed().Number())
	for _, coefficient := range f.coefficients[1:] {
		result.Add(coefficient.Times(accumulator))
		accumulator = ebigint.ToNBigInt(fq.Mul(accumulator.Int, x.Int)).ToRed(x.GetRed())
	}

	return result
}

func (f *FieldVectorPolynomial) InnerProduct(other *FieldVectorPolynomial) []*ebigint.NBigInt {
	b128 := utils.NewBN128()
	var innards = other.GetCoefficients()
	var length = len(f.coefficients) + len(innards) - 1
	var result = make([]*ebigint.NBigInt, length)
	for i := 0; i < length; i++ {
		result[i] = ebigint.ToNBigInt(big.NewInt(0)).ToRed(b128.Q())
	}

	fq := bn128.NewFq(b128.Q().Number())

	for i := 0; i < len(f.coefficients); i++ {
		mine := f.coefficients[i]
		for j := 0; j < len(innards); j++ {
			theirs := innards[j]
			t := mine.InnerProduct(theirs)
			result[i+j] = ebigint.ToNBigInt(fq.Add(result[i+j].Int, t.Int)).ToRed(t.GetRed())
		}
	}

	return result
}

type PedersenCommitment struct {
	params GeneratorParams
	x      *ebigint.NBigInt
	r      *ebigint.NBigInt
}

func NewPedersenCommitment(params GeneratorParams, coefficient *ebigint.NBigInt, b *ebigint.NBigInt) *PedersenCommitment {
	pc := &PedersenCommitment{
		params: params,
		x:      coefficient,
		r:      b,
	}

	return pc
}

func (pc PedersenCommitment) GetX() *ebigint.NBigInt {
	return pc.x
}

func (pc PedersenCommitment) GetR() *ebigint.NBigInt {
	return pc.r
}

func (pc PedersenCommitment) Commit() utils.Point {
	b128 := utils.NewBN128()
	t1 := b128.G1.MulScalar(pc.params.GetG(), pc.x.Int)
	t2 := b128.G1.MulScalar(pc.params.GetH(), pc.r.Int)
	result := b128.G1.Add(t1, t2)

	return result
}

func (pc PedersenCommitment) Add(other *PedersenCommitment) *PedersenCommitment {
	fq := bn128.NewFq(pc.x.GetRed().Number())
	nx := fq.Add(pc.x.Int, other.GetX().Int)
	nr := fq.Add(pc.r.Int, other.GetR().Int)
	result := NewPedersenCommitment(pc.params, ebigint.ToNBigInt(nx).ToRed(other.GetX().GetRed()),
		ebigint.ToNBigInt(nr).ToRed(other.GetR().GetRed()))

	return result
}

func (pc PedersenCommitment) Times(exponent *ebigint.NBigInt) *PedersenCommitment {
	fq := bn128.NewFq(pc.x.GetRed().Number())
	nx := fq.Mul(pc.x.Int, exponent.Int)
	nr := fq.Mul(pc.r.Int, exponent.Int)
	result := NewPedersenCommitment(pc.params, ebigint.ToNBigInt(nx).ToRed(pc.x.GetRed()),
		ebigint.ToNBigInt(nr).ToRed(pc.x.GetRed()))

	return result
}

type PolyCommitment struct {
	coefficientCommitments []*PedersenCommitment
}

func NewPolyCommitment(params GeneratorParams, coefficients []*ebigint.NBigInt) *PolyCommitment {
	b128 := utils.NewBN128()

	pc := &PolyCommitment{}
	pc.coefficientCommitments = make([]*PedersenCommitment, 0)
	tmp := NewPedersenCommitment(params, coefficients[0], ebigint.ToNBigInt(big.NewInt(0)).ToRed(b128.Q()))
	pc.coefficientCommitments = append(pc.coefficientCommitments, tmp)

	for _, coefficient := range coefficients[1:] {
		rand, _ := b128.RanddomScalar()
		npc := NewPedersenCommitment(params, coefficient, rand)
		pc.coefficientCommitments = append(pc.coefficientCommitments, npc)
	}

	return pc
}

func (pc *PolyCommitment) GetCommitments() []utils.Point {
	commitments := make([]utils.Point, len(pc.coefficientCommitments[1:]))
	for i, commitment := range pc.coefficientCommitments[1:] {
		commitments[i] = commitment.Commit()
	}
	return commitments
}

func (pc *PolyCommitment) Evaluate(x *ebigint.NBigInt) *PedersenCommitment {
	var result = pc.coefficientCommitments[0]
	var accumulator = x

	fq := bn128.NewFq(x.GetRed().Number())
	for _, commitment := range pc.coefficientCommitments[1:] {
		result = result.Add(commitment.Times(accumulator))
		accumulator = ebigint.ToNBigInt(fq.Mul(accumulator.Int, x.Int)).ToRed(x.GetRed())
	}
	return result
}

type Polynomial struct {
	coefficients []*ebigint.NBigInt
}

func NewPolynomial(coefficients []*ebigint.NBigInt) *Polynomial {
	poly := &Polynomial{}
	if coefficients != nil && len(coefficients) > 0 {
		poly.coefficients = coefficients
	} else {
		poly.coefficients = make([]*ebigint.NBigInt, 0)
		b128 := utils.NewBN128()
		p := ebigint.ToNBigInt(big.NewInt(1)).ToRed(b128.Q())
		poly.coefficients = append(poly.coefficients, p)
	}

	return poly
}

func (p *Polynomial) Mul(other *Polynomial) *Polynomial {
	b128 := utils.NewBN128()
	fq := bn128.NewFq(b128.Q().Number())
	product := make([]*ebigint.NBigInt, len(p.coefficients))
	for i, b := range p.coefficients {
		product[i] = ebigint.ToNBigInt(fq.Mul(b.Int, other.coefficients[0].Int)).ToRed(b128.Q())
	}

	product = append(product, ebigint.ToNBigInt(big.NewInt(0)).ToRed(b128.Q()))

	if other.coefficients[1].Cmp(big.NewInt(1)) == 0 {
		// product = product.map((product_i, i) => i > 0 ? product_i.redAdd(this.coefficients[i - 1]) : product_i);

		nproduct := make([]*ebigint.NBigInt, len(product))
		for i, b := range product {
			if i > 0 {
				nproduct[i] = ebigint.ToNBigInt(fq.Add(b.Int, product[i-1].Int)).ToRed(b128.Q())
			} else {
				nproduct[i] = b
			}
		}
		product = nproduct
	}

	return NewPolynomial(product)
}
