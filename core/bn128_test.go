package core

import (
	"encoding/hex"
	"github.com/hpb-project/HCash-SDK/common"
	"github.com/hpb-project/HCash-SDK/common/types"
	"github.com/hpb-project/HCash-SDK/core/bn256"
	"github.com/hpb-project/HCash-SDK/core/ebigint"
	solsha3 "github.com/miguelmota/go-solidity-sha3"
	"gotest.tools/assert"
	"math/big"
	"testing"
)

func TestToRed(t *testing.T) {
	q := b128.Q()
	assert.Equal(t, q.Text(16), "30644e72e131a029b85045b68181585d2833e84879b9709143e1f593f0000001")

	var secret = "0x100a1080a8128d4b966bbe15243b9e776db08603f1a36f6d02071fa58d1d5b32"
	x := ebigint.FromBytes(common.FromHex(secret[2:])).ToRed(b128.Q())
	assert.Equal(t, b128.Bytes(x.Int), "0x100a1080a8128d4b966bbe15243b9e776db08603f1a36f6d02071fa58d1d5b32")
}

func TestKeyPair(t *testing.T) {
	var expect_pubkey = types.Point{"0x124c032852ddfcea7e3bdfa7085a8ad013962decab4c230941417d8f859a7e57",
		"0x21af1d2346d59bff8237a442e4464977411496b4bb466a48a058b773874bbea1"}
	var secret = "0x100a1080a8128d4b966bbe15243b9e776db08603f1a36f6d02071fa58d1d5b32"
	x := ebigint.FromBytes(common.FromHex(secret[2:])).ToRed(b128.Q())
	curve_g := b128.Serialize(b128.CurveG())
	assert.Equal(t, curve_g.GX(), "0x077da99d806abd13c9f15ece5398525119d11e11e9836b2ee7d23f6159ad87d4")
	assert.Equal(t, curve_g.GY(), "0x01485efa927f2ad41bff567eec88f32fb0a0f706588b4e41a8d587d008b7f875")

	y := b128.Serialize(b128.CurveG().Mul(x))

	assert.Equal(t, y.GX(), "0x124c032852ddfcea7e3bdfa7085a8ad013962decab4c230941417d8f859a7e57")
	assert.Equal(t, y.GY(), "0x21af1d2346d59bff8237a442e4464977411496b4bb466a48a058b773874bbea1")
	assert.Assert(t, y.Equal(expect_pubkey))
}

func TestBn256(t *testing.T) {
	gX, _ := hex.DecodeString("077da99d806abd13c9f15ece5398525119d11e11e9836b2ee7d23f6159ad87d4")
	gY, _ := hex.DecodeString("01485efa927f2ad41bff567eec88f32fb0a0f706588b4e41a8d587d008b7f875")

	m := BytesCombine(BytePadding(gX, 32), BytePadding(gY, 32))

	g1, ok := new(bn256.G1).Unmarshal(m)
	if !ok {
		t.Error("bn256.G1 unmarshal failed")
		return
	}
	assert.Equal(t, g1.String(), "bn256.G1(77da99d806abd13c9f15ece5398525119d11e11e9836b2ee7d23f6159ad87d4, 1485efa927f2ad41bff567eec88f32fb0a0f706588b4e41a8d587d008b7f875)")

	secret, _ := new(big.Int).SetString("100a1080a8128d4b966bbe15243b9e776db08603f1a36f6d02071fa58d1d5b32", 16)
	pubkey := new(bn256.G1).ScalarMult(g1, secret)
	assert.Equal(t, pubkey.String(), "bn256.G1(124c032852ddfcea7e3bdfa7085a8ad013962decab4c230941417d8f859a7e57, 21af1d2346d59bff8237a442e4464977411496b4bb466a48a058b773874bbea1)")

	secret2, _ := new(big.Int).SetString("760a1080a8128d4b966bbe15243b9e776db08603f1a36f6d02071fa58d1d5882", 16)
	pubkey2 := new(bn256.G1).ScalarMult(g1, secret2)
	assert.Equal(t, pubkey2.String(), "bn256.G1(2dce018002ed0a91922fcb46e9615cc8bfe7ca14aa286d9ec32aa259653c04eb, 2a2af6c7969b6471750a5ae05c93f4bfbfe78381fb586a4d7c9fc1bb2096fff)")

	kadd := new(bn256.G1).Add(pubkey, pubkey2)
	assert.Equal(t, kadd.String(), "bn256.G1(103cd14aad48304f05bcc51b62631240afb108f51568a2dc03200c5745683808, 1f88e682af411144bd770df050537b7a8cd6f81172cea1c560bb36d05458289d)")
}

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
		assert.Equal(t, hash.Text(16), "740e2d6e73d24c31b91f9a93d64eeb3eb5c2df3f0ef3afda1a105605b9cc43e")
	}
}

func TestGEpoch(t *testing.T) {
	epoch := 16
	gepoch := GEpoch(epoch)
	p := b128.Serialize(gepoch)
	assert.Equal(t, p.GX(), "0x24efbd461de73b406c9843a99d04f8212b24a7a9a0c1bb669bf1099e23327501")
	assert.Equal(t, p.GY(), "0x02d2f06f08772767856f3307fe2f19d0a43f370b346f0f08908fa96984584805")
}
