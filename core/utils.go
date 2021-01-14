package core

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	abi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/hpb-project/HCash-SDK/common/types"
	"github.com/hpb-project/HCash-SDK/core/ebigint"
	solsha3 "github.com/miguelmota/go-solidity-sha3"
	"log"
	"math/big"
	"strings"
)

type Account struct {
	X *ebigint.NBigInt `json:"x"`
	Y types.Point      `json:"y"`
}

func (a Account) String() string {
	b, e := json.Marshal(a)
	if e != nil {
		log.Println("account marshal failed, e:", e.Error())
	}
	return string(b)
}

func (a Account) MarshalJSON() ([]byte, error) {
	type pAccount struct {
		X string      `json:"x"`
		Y types.Point `json:"y"`
	}
	var p pAccount
	p.X = a.X.String()
	p.Y = a.Y

	return json.Marshal(p)
}

func (a *Account) UnmarshalJSON(input []byte) error {
	type pAccount struct {
		X string       `json:"x"`
		Y *types.Point `json:"y"`
	}
	var p pAccount
	if err := json.Unmarshal(input, &p); err != nil {
		return err
	}

	nx, ok := new(big.Int).SetString(p.X, 16)
	if !ok {
		return errors.New("invalid hex string for x field")
	}

	a.X = ebigint.ToNBigInt(nx).ForceRed(b128.Q())
	a.Y = *p.Y

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

type ABI_Bytes32 [32]byte
type ETH_ADDR common.Address

func Hash(str string) *ebigint.NBigInt {
	// soliditySha3
	if strings.HasPrefix(str, "0x") { // auto change to u256
		str = str[2:]
	}
	d, _ := hex.DecodeString(str)
	hash := crypto.Keccak256(d)
	//log.Println("keccak256 hash = ", hex.EncodeToString(hash))
	return ebigint.FromBytes(hash).ToRed(b128.Q())
}

// just for test with special k.
func SignWithRandom(address []byte, keypair Account, k *ebigint.NBigInt) (*ebigint.NBigInt, *ebigint.NBigInt, error) {
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
	var addr = ETH_ADDR{}
	copy(addr[:], address[:])

	var bx, by ABI_Bytes32
	bbx, _ := hex.DecodeString(keypair.Y.GX()[2:])
	bby, _ := hex.DecodeString(keypair.Y.GY()[2:])
	copy(bx[:], bbx)
	copy(by[:], bby)

	skey := b128.Serialize(K)
	var kx, ky ABI_Bytes32
	bkx, _ := hex.DecodeString(skey.GX()[2:])
	bky, _ := hex.DecodeString(skey.GY()[2:])
	copy(kx[:], bkx[:])
	copy(ky[:], bky[:])
	thebytes, e := arguments.Pack(
		addr,
		[2]ABI_Bytes32{bx, by},
		[2]ABI_Bytes32{kx, ky},
	)
	if e != nil {
		log.Println("Sign pack failed, e:", e.Error())
		return nil, nil, e
	}
	//log.Println("Sign abiencoder=", hex.EncodeToString(thebytes))

	c := Hash(hex.EncodeToString(thebytes))
	var s = c.RedMul(keypair.X).RedAdd(k)
	//log.Println("Sign c=", c.Text(16))
	//log.Println("Sign s=", s.Text(16))

	return c, s, e
}

func Sign(address []byte, keypair Account) (*ebigint.NBigInt, *ebigint.NBigInt, error) {
	var k = b128.RandomScalar()
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
	var addr = ETH_ADDR{}
	copy(addr[:], address[:])

	var bx, by ABI_Bytes32
	bbx, _ := hex.DecodeString(keypair.Y.GX()[2:])
	bby, _ := hex.DecodeString(keypair.Y.GY()[2:])
	copy(bx[:], bbx)
	copy(by[:], bby)

	skey := b128.Serialize(K)
	var kx, ky ABI_Bytes32
	bkx, _ := hex.DecodeString(skey.GX()[2:])
	bky, _ := hex.DecodeString(skey.GY()[2:])
	copy(kx[:], bkx[:])
	copy(ky[:], bky[:])
	thebytes, e := arguments.Pack(
		addr,
		[2]ABI_Bytes32{bx, by},
		[2]ABI_Bytes32{kx, ky},
	)
	if e != nil {
		log.Println("Sign pack failed, e:", e.Error())
		return nil, nil, e
	}
	//log.Println("Sign abiencoder=", hex.EncodeToString(thebytes))

	c := Hash(hex.EncodeToString(thebytes))
	var s = c.RedMul(keypair.X).RedAdd(k)
	//log.Println("Sign c=", c.Text(16))
	//log.Println("Sign s=", s.Text(16))

	return c, s, nil
}

func CreateAccount() Account {
	x := b128.RandomScalar()
	p := b128.CurveG().Mul(x)
	return Account{X: x, Y: b128.Serialize(p)}
}

func CreateAccountWithX(x *ebigint.NBigInt) Account {
	p := b128.CurveG().Mul(x)
	return Account{X: x, Y: b128.Serialize(p)}
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

	add_one := new(big.Int).Add(b128.P().Int, big.NewInt(1))
	p_1_4 := new(big.Int).Div(add_one, big.NewInt(4))

	for {
		y_squared := seed_red.RedExp(big.NewInt(3)).RedAdd(ebigint.NewNBigInt(3).ToRed(b128.Q()))
		var y = y_squared.RedExp(p_1_4)

		if y.RedExp(big.NewInt(2)).Eq(y_squared) {
			return NewPoint(seed_red.FromRed().Int, y.FromRed().Int)
		}
		seed_red.RedIAdd(ebigint.NewNBigInt(1).ToRed(b128.P()))
	}
}

func GEpoch(epoch int) Point {
	// soliditySha3
	hash := solsha3.SoliditySHA3(solsha3.String("Zether"), solsha3.Uint256(big.NewInt(int64(epoch))))
	hashstr := "0x" + hex.EncodeToString(hash)

	return MapInto(hashstr)
}

func U(epoch int, x *ebigint.NBigInt) Point {
	p := GEpoch(epoch)

	return p.Mul(x)
}
