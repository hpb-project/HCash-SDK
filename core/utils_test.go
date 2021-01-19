package core

import (
	"encoding/hex"
	"github.com/hpb-project/HCash-SDK/common/types"
	"github.com/hpb-project/HCash-SDK/core/ebigint"
	solsha3 "github.com/miguelmota/go-solidity-sha3"
	"gotest.tools/assert"
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
		hexstr := "000000000000000000000000e4920905e06c6b6070477c40b85756ffda3cd3e62bf53727c296faf285546f52437bf8e38752bb5e707124d50b1a3bedb70c99e80bbc4c4de030c083f30eb183ab444e08e18230feb6124d82e6d0677567f9d8e22592890da2861d41f28dfa8cb82e8a39ab24f0028727a4472dd674fa9b5db7912c75c624197b5694da15a93d40cbde84dfba1c46f1a1102ad7fb60970a5ab682"
		hash := Hash(hexstr)
		assert.Equal(t, hash.Text(16), "2262e7e974312cbc363a6cad835026c3c94c3e911ced0a0a380cc273ce9b6d87")
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

func TestCreateAccount(t *testing.T) {
	nx, _ := new(big.Int).SetString("23537dd8704f6cfdfdb0256c3d1c4a6012fb6ae05102762d8d257d5e1ef4fc16", 16)
	account := CreateAccountWithX(ebigint.ToNBigInt(nx))
	assert.Equal(t, account.X.String(), "0x23537dd8704f6cfdfdb0256c3d1c4a6012fb6ae05102762d8d257d5e1ef4fc16")
	assert.Equal(t, account.Y.GX(), "0x012984cced2b6375c23249ea95e451080219a4215b7bfcc20531673d005c8ff0")
	assert.Equal(t, account.Y.GY(), "0x0fb0f2a0c61aca0f0c20f4ac53f55a1f2a8a18b7bc9b3527a4b15173201df29c")
}

func BenchmarkCreateAccount(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		account := CreateAccount()
		account.String()
	}
}

func TestSign(t *testing.T) {
	address, _ := hex.DecodeString("E4920905e06c6B6070477c40B85756ffDa3cD3E6")
	nx, _ := new(big.Int).SetString("299569ae0ae1d40140fd8d9afc54d2f581a292fd13fe88c7033d488119bb95b7", 16)
	account := CreateAccountWithX(ebigint.ToNBigInt(nx))
	//log.Println("test sign with account ", account.String())

	nk, _ := new(big.Int).SetString("2493a56987e869bbb150c14aff5b2e897d9fe78d6dad8b12c92432473f7e9abd", 16)
	sign_k := ebigint.ToNBigInt(nk)

	c, s, e := SignWithRandom(address, account, sign_k)
	if e != nil {
		t.Error("sign failed e:", e.Error())
	}
	assert.Equal(t, PaddingString(c.Text(16), 64), "206db78bfe338ecffd5b2f0606789ff1045bfbf1e46c897f8fa2e2115e19ed74")
	assert.Equal(t, PaddingString(s.Text(16), 64), "003fe7000561eeebccd4bff3160cd7f8fd50db62904d8fa217692a1f6ca8e7ed")
}

func TestReadBalance(t *testing.T) {
	var CL = types.Point{"0x1b5d4b9abe488e61bbb92edff41682560a9d6e02335e2bca9b50881c9540e393", "0x15dc61a9eff5d5a4e70ed97cbce60f7afc69c9925a409ddba365897f1384ca58"}
	var CR = types.Point{"0x0456301d6013d1cc52455a37c8762f2463b1c7e148d55e1c7d9980d8ed8d54b8", "0x27e78199776a73737fa833429fd64e00fa592ca21dda2e92d3489c96148308cb"}
	nx, _ := new(big.Int).SetString("20a89bb465e9e2262e25901525509686f6a26b2fba976f1d9ff00a0cdbb362b0", 16)
	balance := ReadBalance(CL, CR, ebigint.ToNBigInt(nx).ForceRed(b128.Q()))
	assert.Equal(t, balance, 2)
}
