package core

import (
	"encoding/hex"
	"github.com/hpb-project/HCash-SDK/common"
	"github.com/hpb-project/HCash-SDK/common/types"
	"github.com/hpb-project/HCash-SDK/core/bn256"
	"github.com/hpb-project/HCash-SDK/core/ebigint"
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

	kMul := new(bn256.G1).ScalarMult(pubkey, secret)
	assert.Equal(t, kMul.String(), "bn256.G1(2d3f32371440a7b25532c8f5ca3a8d0c52720dafa6d8f17a619a8d7132403602, f5f1a3d3b3407b7ead00192ff0f3836f50a9dd272bc3788b1d733b5a5a0c879)")
}
