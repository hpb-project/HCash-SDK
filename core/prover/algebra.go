package prover

import (
	"encoding/hex"
	"github.com/hpb-project/HCash-SDK/core/ebigint"
	"github.com/hpb-project/HCash-SDK/core/utils"
	"github.com/hpb-project/HCash-SDK/core/utils/bn128"
	solsha3 "github.com/miguelmota/go-solidity-sha3"
	"math/big"
)

type GeneratorParams struct {
	g  utils.Point
	h  utils.Point
	gs GeneratorVector
	hs GeneratorVector
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

func (g GeneratorParams) GetGS() GeneratorVector {
	return g.gs
}

func (g GeneratorParams) GetHS() GeneratorVector {
	return g.hs
}

func (g *GeneratorParams) Commit(blinding, gExp, hExp string) {
	// todo: implement commit.
}

type GeneratorVector struct {
}

func NewGeneratorVector(Innards []utils.Point) GeneratorVector {
	gv := &GeneratorVector{}

	return *gv
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

func (c *Convolver) FFT(input, inverse []string) []string {
	var length = len(input)
	if length == 1 {
		return input
	}
	if length%2 != 0 {
		panic("Input size must be a power of 2!")
	}
	fq := bn128.NewFq(c.unity.GetRed().Number())
	base := big.NewInt(1).Lsh(big.NewInt(1), 28)
	var omegas = fq.Exp(c.unity.Int, base.Div(base, big.NewInt(int64(length))))
	if inverse != nil && len(inverse) > 0 {
		fq.Inverse()
	}
}

type FieldVectorPolynomial struct {
	coefficients []*PedersenCommitment
}

func NewFieldVectorPolynomial(coefficients []*PedersenCommitment) *FieldVectorPolynomial {
	fvp := &FieldVectorPolynomial{
		coefficients: coefficients,
	}
	return fvp
}

func (f *FieldVectorPolynomial) GetCoefficients() []*PedersenCommitment {
	return f.coefficients
}

func (f *FieldVectorPolynomial) Evaluate(x *ebigint.NBigInt) *PedersenCommitment {
	result := f.coefficients[0]
	var accumulator = x
	fq := bn128.NewFq(x.GetRed().Number())
	for _, coefficient := range f.coefficients[1:] {
		result.Add(coefficient.Times(accumulator))
		accumulator = ebigint.ToNBigInt(fq.Mul(accumulator.Int, x.Int))
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
			result[i+j] = ebigint.ToNBigInt(fq.Add(result[i+j].Int, mine.InnerProduct(theirs)))
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
	bn128 := utils.NewBN128()
	t1 := bn128.G1.MulScalar(pc.params.GetG(), pc.x.Int)
	t2 := bn128.G1.MulScalar(pc.params.GetH(), pc.r.Int)
	result := bn128.G1.Add(t1, t2)

	return utils.Point(result)
}

func (pc PedersenCommitment) Add(other *PedersenCommitment) *PedersenCommitment {
	fq := bn128.NewFq(pc.x.GetRed().Number())
	nx := fq.Add(pc.x.Int, other.GetX().Int)
	nr := fq.Add(pc.r.Int, other.GetR().Int)
	result := NewPedersenCommitment(pc.params, ebigint.ToNBigInt(nx), ebigint.ToNBigInt(nr))

	return result
}

func (pc PedersenCommitment) Times(exponent *ebigint.NBigInt) *PedersenCommitment {
	fq := bn128.NewFq(pc.x.GetRed().Number())
	nx := fq.Mul(pc.x.Int, exponent.Int)
	nr := fq.Mul(pc.r.Int, exponent.Int)
	result := NewPedersenCommitment(pc.params, ebigint.ToNBigInt(nx), ebigint.ToNBigInt(nr))

	return result
}

type PolyCommitment struct {
	params                 GeneratorParams
	coefficientCommitments []*PedersenCommitment
}

func NewPolyCommitment(params GeneratorParams, coefficients []*ebigint.NBigInt) *PolyCommitment {
	bn128 := utils.NewBN128()

	pc := &PolyCommitment{}
	pc.params = params
	pc.coefficientCommitments = make([]*PedersenCommitment, 0)
	tmp := NewPedersenCommitment(params, coefficients[0], ebigint.ToNBigInt(big.NewInt(0)).ToRed(bn128.Q()))
	pc.coefficientCommitments = append(pc.coefficientCommitments, tmp)

	for _, coefficient := range coefficients[1:] {
		rand, _ := bn128.RanddomScalar()
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
		accumulator = ebigint.ToNBigInt(fq.Mul(accumulator.Int, x.Int))
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
		bn128 := utils.NewBN128()
		p := ebigint.ToNBigInt(big.NewInt(1)).ToRed(bn128.Q())
		poly.coefficients = []*ebigint.NBigInt{p}
	}

	return poly
}

func (p *Polynomial) Mul(other *Polynomial) *Polynomial {
	fq := bn128.NewFq(utils.NewBN128().Q().Number())
	product := make([]*ebigint.NBigInt, len(p.coefficients))
	for i, b := range p.coefficients {
		product[i] = ebigint.ToNBigInt(fq.Mul(b.Int, other.coefficients[0].Int))
	}

	product = append(product, ebigint.ToNBigInt(big.NewInt(0)).ToRed(utils.NewBN128().Q()))

	if other.coefficients[1].Cmp(big.NewInt(1)) == 0 {
		// product = product.map((product_i, i) => i > 0 ? product_i.redAdd(this.coefficients[i - 1]) : product_i);

		nproduct := make([]*ebigint.NBigInt, len(product))
		for i, b := range product {
			if i > 0 {
				nproduct[i] = ebigint.ToNBigInt(fq.Add(b.Int, product[i-1].Int))
			} else {
				nproduct[i] = b
			}
		}
		product = nproduct
	}

	return NewPolynomial(product)
}
