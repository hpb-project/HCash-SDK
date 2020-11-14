package utils

import (
	"encoding/hex"
	abi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hpb-project/HCash-SDK/core/ebigint"
	bn1282 "github.com/hpb-project/HCash-SDK/core/utils/bn128"
	solsha3 "github.com/miguelmota/go-solidity-sha3"
	"math/big"
)

type Pubkey struct {
	gX string
	gY string
}

func ReadBalance(CL, CR Pubkey, x *ebigint.NBigInt) int {
	bn128 := NewBN128()
	nCL := bn128.UnSerialize(CL.gX, CL.gY)
	nCR := bn128.UnSerialize(CR.gX, CR.gY)

	neg := bn128.FQ().Neg(x.Int)
	tmp := bn128.G1.MulScalar(nCR, neg)
	var gB = bn128.G1.Add(nCL, tmp)
	var accumulator = bn128.Zero()

	for i := 0; i < bn128.B_MAX(); i++ {
		if bn128.G1.Equal(accumulator, gB) {
			return i
		}
		accumulator = bn128.G1.Add(accumulator, bn128.G1.G)
	}

	return 0
}

func Hash(str string) *ebigint.NBigInt {
	bn128 := NewBN128()

	// soliditySha3
	hash := solsha3.SoliditySHA3(solsha3.String(str[2:]))
	hexstr := hex.EncodeToString(hash)
	n, _ := big.NewInt(0).SetString(hexstr, 16)
	return ebigint.ToNBigInt(n).ToRed(bn128.Q())
}

func Sign(address string, keypair []string) []string {
	bn128 := NewBN128()
	var k, _ = bn128.RanddomScalar()
	K := bn128.G1.MulScalar(bn128.G1.G, k.Int)
	sx, sy := bn128.Serialize(K)

	addressT, _ := abi.NewType("address", "", nil)
	bytes32_2T, _ := abi.NewType("bytes32[2]", "", nil)

	arguments := abi.Arguments{
		{
			Type: addressT,
		},
		{
			Type: bytes32_2T,
		},
		{
			Type: bytes32_2T,
		},
	}
	bytes, _ := arguments.Pack(
		common.HexToAddress(address),
		[2]string{keypair[1], keypair[2]},
		[2]string{sx, sy},
	)

	bstr := hex.EncodeToString(bytes)

	c := Hash(bstr)

	privk, _ := big.NewInt(0).SetString(keypair[0], 16)
	p := bn128.FQ().Mul(c.Int, privk)
	s := bn128.FQ().Add(p, k.Int)

	return []string{bn128.Bytes(c.Int), bn128.Bytes(s)}
}

func CreateAccount() (string, string, string) {
	b128 := NewBN128()
	x, _ := b128.RanddomScalar()
	p := b128.G1.MulScalar(b128.G1.G, x.Int)
	priv := b128.Bytes(x.Int)
	pub_x, pub_y := b128.Serialize(Point{p[0], p[1], p[2]})

	return priv, pub_x, pub_y
}

//
//utils.mapInto = (seed) => { // seed is flattened 0x + hex string
//		var seed_red = new BN(seed.slice(2), 16).toRed(bn128.p);
//		var p_1_4 = bn128.curve.p.add(new BN(1)).div(new BN(4));
//		while (true) {
//			var y_squared = seed_red.redPow(new BN(3)).redAdd(new BN(3).toRed(bn128.p));
//			var y = y_squared.redPow(p_1_4);
//			if (y.redPow(new BN(2)).eq(y_squared)) {
//				return bn128.curve.point(seed_red.fromRed(), y.fromRed());
//			}
//			seed_red.redIAdd(new BN(1).toRed(bn128.p));
//		}
//};

func MapInto(seed string) Point {
	bn128 := NewBN128()
	fq := bn1282.NewFq(bn128.p)
	n, _ := big.NewInt(0).SetString(seed[2:], 16)
	seed_red := ebigint.ToNBigInt(n).ToRed(bn128.P())
	one := big.NewInt(1)
	p1_4 := one.Div(one, big.NewInt(4))
	p_1_4 := bn128.p.Add(bn128.p, p1_4)

	for {
		y_squared := ebigint.ToNBigInt(fq.Add(bn128.FQ().Exp(seed_red.Int, big.NewInt(3)),
			big.NewInt(3))).ToRed(bn128.P())
		y := fq.Exp(y_squared.Int, p_1_4)

		if fq.Equal(y_squared.Int, fq.Exp(y, big.NewInt(2))) {
			return NewPoint(bn128, seed_red, ebigint.ToNBigInt(y).ToRed(bn128.P()))
		}
		fq.Add(seed_red.Int, ebigint.ToNBigInt(big.NewInt(1)).ToRed(bn128.P()).Int)
	}

}

func GEpoch(epoch uint) Point {

	// soliditySha3
	// todo : change the type of epoch with contract defined.
	hash := solsha3.SoliditySHA3(solsha3.String("Zether"),
		solsha3.Uint32(epoch))
	hashstr := "0x" + hex.EncodeToString(hash)

	return MapInto(hashstr)
}

func U(epoch uint, x *big.Int) Point {
	bn128 := NewBN128()
	p := GEpoch(epoch)

	return bn128.G1.MulScalar(p, x)
}
