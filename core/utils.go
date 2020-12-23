package core

import (
	"encoding/hex"
	"encoding/json"
	abi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/hpb-project/HCash-SDK/common/types"
	"github.com/hpb-project/HCash-SDK/core/ebigint"
	solsha3 "github.com/miguelmota/go-solidity-sha3"
	"math/big"
)

type Account struct {
	X *ebigint.NBigInt `json:"x"`
	Y types.Point      `json:"y"`
}

func (a Account) String() string {
	b, _ := json.Marshal(a)
	return string(b)
}

func (a Account) MarshalJSON() ([]byte, error) {
	type pAccount struct {
		x *ebigint.NBigInt `json:"x"`
		y types.Point      `json:"y"`
	}
	var p pAccount
	p.x = a.X
	p.y = a.Y

	return json.Marshal(p)
}

func (a Account) UnmarshalJSON(input []byte) error {
	type pAccount struct {
		x *ebigint.NBigInt `json:"x"`
		y types.Point      `json:"y"`
	}
	var p pAccount
	if err := json.Unmarshal(input, &p); err != nil {
		return err
	}
	a.X = p.x.ForceRed(b128.Q())
	a.Y = p.y

	return nil
}

func ReadBalance(CL, CR types.Point, x *ebigint.NBigInt) int {
	nCL := b128.UnSerialize(CL)
	nCR := b128.UnSerialize(CR)

	var gB = nCL.Add(nCR.Mul(x.RedNeg()))
	var accumulator = b128.Zero()

	for i := 0; i < b128.B_MAX(); i++ {
		if accumulator.Equal(gB) {
			return i
		}
		accumulator = accumulator.Add(b128.CurveG())
	}

	return 0
}

func Hash(str string) *ebigint.NBigInt {
	// soliditySha3
	hash := solsha3.SoliditySHA3(solsha3.String(str[2:]))
	hexstr := hex.EncodeToString(hash)
	n, _ := big.NewInt(0).SetString(hexstr, 16)
	return ebigint.ToNBigInt(n).ToRed(b128.Q())
}

func Sign(address []byte, keypair Account) string {
	var k = b128.RanddomScalar()
	var K = b128.CurveG().Mul(k)

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

	skey := b128.Serialize(K)
	bytes, _ := arguments.Pack(
		address,
		keypair.Y,
		[2]string(skey),
	)

	c := Hash(hex.EncodeToString(bytes))
	var s = c.RedMul(keypair.X).RedAdd(k)
	type CS struct {
		C string `json:"c"`
		S string `json:"s"`
	}
	var ret_cs = CS{
		C: b128.Bytes(c.Int),
		S: b128.Bytes(s.Int),
	}
	data, _ := json.Marshal(ret_cs)
	return string(data)
}

func CreateAccount() Account {
	x := b128.RanddomScalar()
	p := b128.CurveG().Mul(x)

	return Account{
		X: x,
		Y: b128.Serialize(p),
	}
}

/*
utils.mapInto = (seed) => { // seed is flattened 0x + hex string
    var seed_red = new BN(seed.slice(2), 16).toRed(bn128.p);
    var p_1_4 = bn128.curve.p.add(new BN(1)).div(new BN(4));
    while (true) {
        var y_squared = seed_red.redPow(new BN(3)).redAdd(new BN(3).toRed(bn128.p));
        var y = y_squared.redPow(p_1_4);
        if (y.redPow(new BN(2)).eq(y_squared)) {
            return bn128.curve.point(seed_red.fromRed(), y.fromRed());
        }
        seed_red.redIAdd(new BN(1).toRed(bn128.p));
    }
};
*/
// seed is flattened 0x + hex string
func MapInto(seed string) Point {
	n, _ := big.NewInt(0).SetString(seed[2:], 16)
	seed_red := ebigint.ToNBigInt(n).ToRed(b128.P())

	one := big.NewInt(1)
	p1_4 := one.Div(one, big.NewInt(4))
	p_1_4 := one.Add(b128.P().Int, p1_4)

	for {
		y_squared := seed_red.RedExp(big.NewInt(3)).RedAdd(ebigint.NewNBigInt(3).ToRed(b128.Q()))
		var y = y_squared.RedExp(p_1_4)

		if y.RedExp(big.NewInt(2)).Eq(y_squared) {
			return NewPoint(b128, seed_red.FromRed().Int, y.FromRed().Int)
		}
		seed_red.RedIAdd(ebigint.NewNBigInt(1).ToRed(b128.P()))
	}
}

func GEpoch(epoch int) Point {

	// soliditySha3
	// todo : change the type of epoch with contract defined.
	hash := solsha3.SoliditySHA3(solsha3.String("Zether"),
		solsha3.Uint32(epoch))
	hashstr := "0x" + hex.EncodeToString(hash)

	return MapInto(hashstr)
}

func U(epoch int, x *ebigint.NBigInt) Point {
	p := GEpoch(epoch)

	return p.Mul(x)
}
