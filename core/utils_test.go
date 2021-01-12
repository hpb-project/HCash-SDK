package core

import (
	"encoding/hex"
	"github.com/hpb-project/HCash-SDK/core/ebigint"
	solsha3 "github.com/miguelmota/go-solidity-sha3"
	"gotest.tools/assert"
	"log"
	"math/big"
	"testing"
)

func TestSoliditySha3(t *testing.T) {
	hash3 := solsha3.SoliditySHA3(
		solsha3.String("hello"),
	)
	str := hex.EncodeToString(hash3)
	assert.Equal(t, str, "1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8")

	hash := solsha3.SoliditySHA3(solsha3.String("Zether"), solsha3.Uint256(big.NewInt(16)))
	str = hex.EncodeToString(hash)
	assert.Equal(t, str, "85b85a2be04a7b93dd38cf16a007a8dc5a277ccc71a55081143221cbd42c6f8f")
}

func TestHash(t *testing.T) {
	{
		hexstr := "0x85b85a2be04a7b93dd38cf16a007a8dc5a277ccc71a55081143221cbd42c6f8f"
		hash := Hash(hexstr)
		assert.Equal(t, hash.Text(16), "1dad24f83633aebe1d4742c6a843fb82dc5e84dbcbca2c5af2dbccd6af8a2700")
	}
	{
		str := "85b85a2be04a7b93dd38cf16a007a8dc5a277ccc71a55081143221cbd42c6f8f"
		hash := Hash(str)
		assert.Equal(t, hash.Text(16), "1dad24f83633aebe1d4742c6a843fb82dc5e84dbcbca2c5af2dbccd6af8a2700")
	}
}

func TestGEpoch(t *testing.T) {
	epoch := 16
	gepoch := GEpoch(epoch)
	p := b128.Serialize(gepoch)
	assert.Equal(t, p.GX(), "0x24efbd461de73b406c9843a99d04f8212b24a7a9a0c1bb669bf1099e23327501")
	assert.Equal(t, p.GY(), "0x02d2f06f08772767856f3307fe2f19d0a43f370b346f0f08908fa96984584805")
}

func TestU(t *testing.T) {
	epoch := 16
	x, _ := new(big.Int).SetString("23537dd8704f6cfdfdb0256c3d1c4a6012fb6ae05102762d8d257d5e1ef4fc16", 16)

	P := U(epoch, ebigint.ToNBigInt(x))
	sp := b128.Serialize(P)
	assert.Equal(t, sp.GX(), "0x171ea019e27e1c83e5faa817d93324aeabb3c33beda426d34430684da598526b")
	assert.Equal(t, sp.GY(), "0x125a81290d856b8be5c85903e09b22652b3c289e667e81c70f60a31801c6b6f1")
}

func TestSign(t *testing.T) {
	address, _ := hex.DecodeString("E4920905e06c6B6070477c40B85756ffDa3cD3E6")
	nx, _ := new(big.Int).SetString("23537dd8704f6cfdfdb0256c3d1c4a6012fb6ae05102762d8d257d5e1ef4fc16", 16)
	account := CreateAccountWithX(ebigint.ToNBigInt(nx))
	log.Println("create account = ", account.String())

	nk, _ := new(big.Int).SetString("2493a56987e869bbb150c14aff5b2e897d9fe78d6dad8b12c92432473f7e9abd", 16)
	sign_k := ebigint.ToNBigInt(nk)

	result := SignWithRandom(address, account, sign_k)
	log.Println("sign result = ", result)
}
