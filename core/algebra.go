package core

import (
	"encoding/hex"
	"reflect"

	//"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/hpb-project/HCash-SDK/core/ebigint"
	solsha3 "github.com/miguelmota/go-solidity-sha3"
	"math/big"
)

type GeneratorParams struct {
	g  Point
	h  Point
	gs *GeneratorVector
	hs *GeneratorVector
}

func NewGeneratorParams(hi interface{}, gs, hs *GeneratorVector) *GeneratorParams {
	gp := &GeneratorParams{}
	hash := solsha3.SoliditySHA3(solsha3.String("G"))
	gp.g = MapInto(hex.EncodeToString(hash))

	h_types := reflect.TypeOf(hi).String()
	if h_types == "int" {
		gsInnards := make([]Point, 0)
		hsInnards := make([]Point, 0)
		hVal := hi.(int)
		for i := 0; i < hVal; i++ {
			hash1 := solsha3.SoliditySHA3(solsha3.String("G"), solsha3.Int256(i))
			p1 := MapInto(hex.EncodeToString(hash1))
			gsInnards = append(gsInnards, p1)

			hash2 := solsha3.SoliditySHA3(
				solsha3.String("H"), solsha3.Int256(i))
			p2 := MapInto(hex.EncodeToString(hash2))
			hsInnards = append(hsInnards, p2)
		}
		gp.h = MapInto(hex.EncodeToString(solsha3.SoliditySHA3(solsha3.String("H"))))

		gp.gs = NewGeneratorVector(gsInnards)
		gp.hs = NewGeneratorVector(hsInnards)
	} else {
		gp.h = hi.(Point)
		gp.gs = gs
		gp.hs = hs
	}

	return gp
}

func (g GeneratorParams) GetG() Point {
	return g.g
}

func (g GeneratorParams) GetH() Point {
	return g.h
}

func (g GeneratorParams) GetGS() *GeneratorVector {
	return g.gs
}

func (g GeneratorParams) GetHS() *GeneratorVector {
	return g.hs
}

func (g *GeneratorParams) Commit(blinding *ebigint.NBigInt, gExp, hExp *FieldVector) Point {
	var result = g.h.Mul(blinding)
	var gsVector = g.gs.GetVector()
	gexpVector := gExp.GetVector()
	for i, gexp := range gexpVector {
		result = result.Add(gsVector[i].Mul(gexp))
	}

	if hExp != nil {
		var hsVector = g.hs.GetVector()
		hexpVector := hExp.GetVector()
		for i, hexp := range hexpVector {
			result = result.Add(hsVector[i].Mul(hexp))
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
	var innards = other.GetVector()
	var nInnards = make([]*ebigint.NBigInt, len(f.vector))

	for i, elem := range f.vector {
		nInnards[i] = elem.RedAdd(innards[i])
	}

	return NewFieldVector(nInnards)
}

func (f *FieldVector) Plus(constant *ebigint.NBigInt) *FieldVector {

	var nInnards = make([]*ebigint.NBigInt, len(f.vector))
	for i, elem := range f.vector {
		nInnards[i] = elem.RedAdd(constant)
	}

	return NewFieldVector(nInnards)
}

func (f *FieldVector) Sum() *ebigint.NBigInt {
	var nVectors = make([]*ebigint.NBigInt, 0)

	for _, c := range f.vector {
		nVectors = append(nVectors, c)
	}

	var accumulator = ebigint.NewNBigInt(0).ToRed(b128.Q())
	for _, current := range nVectors {
		accumulator = accumulator.RedAdd(current)
	}

	return accumulator
}

func (f *FieldVector) Negate() *FieldVector {
	var nInnards = make([]*ebigint.NBigInt, len(f.vector))
	for i, accum := range f.vector {
		nInnards[i] = accum.RedNeg()
	}

	return NewFieldVector(nInnards)
}

func (f *FieldVector) Subtract(other *FieldVector) *FieldVector {
	return f.Add(other.Negate())
}

func (f *FieldVector) Hadamard(other *FieldVector) *FieldVector {
	var innards = other.GetVector()
	var nInnards = make([]*ebigint.NBigInt, len(f.vector))

	for i, elem := range f.vector {
		nInnards[i] = elem.RedMul(innards[i])
	}

	return NewFieldVector(nInnards)
}

func (f *FieldVector) Invert() *FieldVector {
	var nInnards = make([]*ebigint.NBigInt, len(f.vector))
	for i, elem := range f.vector {
		nInnards[i] = elem.RedInvm()
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
	var size = f.Length()
	var nInnards = make([]*ebigint.NBigInt, size)
	for i, _ := range nInnards {
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
	var nInnards = make([]*ebigint.NBigInt, len(f.vector))
	for i, elem := range f.vector {
		nInnards[i] = elem.RedMul(constant)
	}

	return NewFieldVector(nInnards)
}

func (f *FieldVector) InnerProduct(other *FieldVector) *ebigint.NBigInt {
	var innards = other.GetVector()
	var nVectors = make([]*ebigint.NBigInt, 0)

	for _, c := range f.vector {
		nVectors = append(nVectors, c)
	}

	var accumulator = ebigint.ToNBigInt(big.NewInt(0)).ToRed(b128.Q())

	for i, current := range nVectors {
		accumulator = accumulator.RedAdd(current.RedMul(innards[i]))
	}

	return accumulator
}

type GeneratorVector struct {
	vector []Point
}

func NewGeneratorVector(Innards []Point) *GeneratorVector {
	gv := &GeneratorVector{}
	gv.vector = Innards
	return gv
}

func (g *GeneratorVector) GetVector() []Point {
	return g.vector
}

func (g *GeneratorVector) Length() int {
	return len(g.vector)
}

func (g *GeneratorVector) Slice(begin, end int) *GeneratorVector {
	return NewGeneratorVector(g.vector[begin:end])
}

func (g *GeneratorVector) Commit(exponents *FieldVector) Point {
	var nVectors = make([]Point, 0)
	var innards = exponents.GetVector()

	for _, c := range g.vector {
		nVectors = append(nVectors, c)
	}

	var accumulator = b128.Zero()

	for i, current := range nVectors {
		accumulator = accumulator.Add(current.Mul(innards[i]))
	}

	return accumulator
}

func (g *GeneratorVector) Sum() Point {
	var nVectors = make([]Point, 0)

	for _, c := range g.vector {
		nVectors = append(nVectors, c)
	}

	var accumulator = b128.Zero()

	for _, current := range nVectors {
		accumulator = accumulator.Add(current)
	}

	return accumulator
}

func (g *GeneratorVector) Add(other *GeneratorVector) *GeneratorVector {
	var innards = other.GetVector()
	var nInnards = make([]Point, len(g.vector))
	for i, elem := range g.vector {
		nInnards[i] = elem.Add(innards[i])
	}

	return NewGeneratorVector(nInnards)
}

func (g *GeneratorVector) Hadamard(exponents *FieldVector) *GeneratorVector {
	var innards = exponents.GetVector()
	var nInnards = make([]Point, len(g.vector))

	for i, elem := range g.vector {
		nInnards[i] = elem.Mul(innards[i])
	}

	return NewGeneratorVector(nInnards)
}

func (g *GeneratorVector) Negate() *GeneratorVector {
	var nInnards = make([]Point, len(g.vector))
	for i, elem := range g.vector {
		nInnards[i] = elem.Neg()
	}

	return NewGeneratorVector(nInnards)
}

func (g *GeneratorVector) Times(constant *ebigint.NBigInt) *GeneratorVector {
	var nInnards = make([]Point, len(g.vector))

	for i, elem := range g.vector {
		nInnards[i] = elem.Mul(constant)
	}

	return NewGeneratorVector(nInnards)
}

func (g *GeneratorVector) Extract(parity int) *GeneratorVector {
	var nInnards = make([]Point, len(g.vector))
	for i, elem := range g.vector {
		if i%2 == parity {
			nInnards = append(nInnards, elem)
		}
	}

	return NewGeneratorVector(nInnards)
}

func (g *GeneratorVector) Concat(other *GeneratorVector) *GeneratorVector {
	var nInnards = make([]Point, 0)
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

	unity, _ := big.NewInt(0).SetString("14a3074b02521e3b1ed9852e5028452693e87be4e910500c7ba9bbddb2f46edd", 16)
	c.unity = ebigint.ToNBigInt(unity).ToRed(b128.Q())

	return c
}

func (c *Convolver) FFT_Scalar(input *FieldVector, inverse bool) *FieldVector {
	var length = input.Length()
	if length == 1 {
		return input
	}
	if length%2 != 0 {
		panic("Input size must be a power of 2!")
	}

	exp := big.NewInt(1).Lsh(big.NewInt(1), 28)
	exp = exp.Div(exp, big.NewInt(int64(length)))

	var omega = c.unity.RedExp(exp)
	if inverse {
		omega = omega.RedInvm()
	}
	var even = c.FFT_Scalar(input.Extract(0), inverse)
	var odd = c.FFT_Scalar(input.Extract(1), inverse)

	var omegas = make([]*ebigint.NBigInt, 0)
	omegas = append(omegas, ebigint.NewNBigInt(1).ToRed(b128.Q()))

	for i := 1; i < length/2; i++ {
		omegas = append(omegas, omegas[i-1].RedMul(omega))
	}

	var n_omegas = NewFieldVector(omegas)
	var result = even.Add(odd.Hadamard(n_omegas)).Concat(even.Add(odd.Hadamard(n_omegas).Negate()))
	if inverse {
		result = result.Times(ebigint.NewNBigInt(2).ToRed(b128.Q()).RedInvm())
	}
	return result
}

func (c *Convolver) Convolution_Scalar(exponent *FieldVector, base *FieldVector) *FieldVector {
	size := base.Length()
	temp := c.FFT_Scalar(base, false).Hadamard(c.FFT_Scalar(exponent.Flip(), false))

	return c.FFT_Scalar(temp.Slice(0, size/2).Add(temp.Slice(size/2, size)).Times(ebigint.NewNBigInt(2).ToRed(b128.Q()).RedInvm()), true)
}

func (c *Convolver) Convolution_Point(exponent *FieldVector, base *GeneratorVector) *GeneratorVector {
	size := base.Length()
	temp := c.FFT_Point(base, false).Hadamard(c.FFT_Scalar(exponent.Flip(), false))

	return c.FFT_Point(temp.Slice(0, size/2).Add(temp.Slice(size/2, size)).Times(ebigint.NewNBigInt(2).ToRed(b128.Q()).RedInvm()), true)
}

func (c *Convolver) FFT_Point(input *GeneratorVector, inverse bool) *GeneratorVector {
	var length = input.Length()
	if length == 1 {
		return input
	}
	if length%2 != 0 {
		panic("Input size must be a power of 2!")
	}
	exp := big.NewInt(1).Lsh(big.NewInt(1), 28)
	exp = exp.Div(exp, big.NewInt(int64(length)))
	var omega = c.unity.RedExp(exp)
	if inverse {
		omega = omega.RedInvm()
	}
	var even = c.FFT_Point(input.Extract(0), inverse)
	var odd = c.FFT_Point(input.Extract(1), inverse)

	var omegas = make([]*ebigint.NBigInt, 0)
	omegas = append(omegas, ebigint.NewNBigInt(1).ToRed(b128.Q()))

	for i := 1; i < length/2; i++ {
		omegas = append(omegas, omegas[i-1].RedMul(omega))
	}

	var n_omegas = NewFieldVector(omegas)
	var result = even.Add(odd.Hadamard(n_omegas)).Concat(even.Add(odd.Hadamard(n_omegas).Negate()))
	if inverse {
		result = result.Times(ebigint.NewNBigInt(2).ToRed(b128.Q()).RedInvm())
	}
	return result
}

type FieldVectorPolynomial struct {
	//coefficients []*PedersenCommitment
	coefficients []*FieldVector
}

func NewFieldVectorPolynomial(coefficients ...*FieldVector) *FieldVectorPolynomial {
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

	for _, coefficient := range f.coefficients[1:] {
		result = result.Add(coefficient.Times(accumulator))
		accumulator = accumulator.RedMul(x)
	}

	return result
}

func (f *FieldVectorPolynomial) InnerProduct(other *FieldVectorPolynomial) []*ebigint.NBigInt {
	var innards = other.GetCoefficients()
	var length = len(f.coefficients) + len(innards) - 1
	var result = make([]*ebigint.NBigInt, length)
	for i := 0; i < length; i++ {
		result[i] = ebigint.NewNBigInt(0).ToRed(b128.Q())
	}

	for i, mine := range f.coefficients {
		for j, their := range innards {
			result[i+j] = result[i+j].RedAdd(mine.InnerProduct(their))
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

func (pc PedersenCommitment) Commit() Point {
	return pc.params.GetG().Mul(pc.x).Add(pc.params.GetH().Mul(pc.r))
}

func (pc PedersenCommitment) Add(other *PedersenCommitment) *PedersenCommitment {
	return NewPedersenCommitment(pc.params, pc.x.RedAdd(other.GetX()), pc.r.RedAdd(other.GetR()))
}

func (pc PedersenCommitment) Times(exponent *ebigint.NBigInt) *PedersenCommitment {
	return NewPedersenCommitment(pc.params, pc.x.RedMul(exponent), pc.r.RedMul(exponent))
}

type PolyCommitment struct {
	coefficientCommitments []*PedersenCommitment
}

func NewPolyCommitment(params GeneratorParams, coefficients []*ebigint.NBigInt) *PolyCommitment {
	pc := &PolyCommitment{}
	pc.coefficientCommitments = make([]*PedersenCommitment, 0)
	tmp := NewPedersenCommitment(params, coefficients[0], ebigint.NewNBigInt(0).ToRed(b128.Q()))
	pc.coefficientCommitments = append(pc.coefficientCommitments, tmp)

	for _, coefficient := range coefficients[1:] {
		rand := b128.RandomScalar()
		npc := NewPedersenCommitment(params, coefficient, rand)
		pc.coefficientCommitments = append(pc.coefficientCommitments, npc)
	}

	return pc
}

func (pc *PolyCommitment) GetCommitments() []Point {
	commitments := make([]Point, len(pc.coefficientCommitments[1:]))
	for i, commitment := range pc.coefficientCommitments[1:] {
		commitments[i] = commitment.Commit()
	}
	return commitments
}

func (pc *PolyCommitment) Evaluate(x *ebigint.NBigInt) *PedersenCommitment {
	var result = pc.coefficientCommitments[0]
	var accumulator = x

	for _, commitment := range pc.coefficientCommitments[1:] {
		result = result.Add(commitment.Times(accumulator))
		accumulator = accumulator.RedMul(x)
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
		poly.coefficients = append(poly.coefficients, ebigint.NewNBigInt(1).ToRed(b128.Q()))
	}

	return poly
}

func (p *Polynomial) Mul(other *Polynomial) *Polynomial {

	product := make([]*ebigint.NBigInt, len(p.coefficients))
	for i, b := range p.coefficients {
		product[i] = b.RedMul(other.coefficients[0])
	}

	product = append(product, ebigint.NewNBigInt(0).ToRed(b128.Q()))

	if other.coefficients[1].Cmp(big.NewInt(1)) == 0 {
		// product = product.map((product_i, i) => i > 0 ? product_i.redAdd(this.coefficients[i - 1]) : product_i);

		nproduct := make([]*ebigint.NBigInt, len(product))
		for i, b := range product {
			if i > 0 {
				nproduct[i] = b.RedAdd(p.coefficients[i-1])
			} else {
				nproduct[i] = b
			}
		}
		product = nproduct
	}

	return NewPolynomial(product)
}
