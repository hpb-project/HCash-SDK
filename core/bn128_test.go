package core

import (
	"encoding/hex"
	"github.com/hpb-project/HCash-SDK/common"
	"github.com/hpb-project/HCash-SDK/core/bn256"
	"github.com/hpb-project/HCash-SDK/core/ebigint"
	"log"
	"math/big"
	"testing"
)

func TestToRed(t *testing.T) {
	q := b128.Q()
	if q.Text(16) != "30644e72e131a029b85045b68181585d2833e84879b9709143e1f593f0000001" {
		t.Error("b128.q is not match")
	}

	var secret = "0x100a1080a8128d4b966bbe15243b9e776db08603f1a36f6d02071fa58d1d5b32"
	x := ebigint.FromBytes(common.FromHex(secret[2:])).ToRed(b128.Q())
	hex_x := b128.Bytes(x.Int)
	if hex_x != "0x100a1080a8128d4b966bbe15243b9e776db08603f1a36f6d02071fa58d1d5b32" {
		t.Error("tored is not match")
	}
}

func TestKeyPair(t *testing.T) {
	var secret = "0x100a1080a8128d4b966bbe15243b9e776db08603f1a36f6d02071fa58d1d5b32"
	x := ebigint.FromBytes(common.FromHex(secret[2:])).ToRed(b128.Q())
	curve_g := b128.Serialize(b128.CurveG())
	log.Println("curve g=", curve_g)
	y := b128.Serialize(b128.CurveG().Mul(x))
	log.Println("y=", y)
}

func TestBn256(t *testing.T) {
	log.Println("in test bn256")
	gX, _ := hex.DecodeString("077da99d806abd13c9f15ece5398525119d11e11e9836b2ee7d23f6159ad87d4")
	gY, _ := hex.DecodeString("01485efa927f2ad41bff567eec88f32fb0a0f706588b4e41a8d587d008b7f875")

	m := BytesCombine(BytePadding(gX, 32), BytePadding(gY, 32))
	log.Println("m= ", hex.EncodeToString(m), "m length = ", len(m))

	g1, ok := new(bn256.G1).Unmarshal(m)
	if !ok {
		log.Println("unmarshal failed")
		//return
	}
	log.Println("g1:", g1.String())

	secret, _ := new(big.Int).SetString("100a1080a8128d4b966bbe15243b9e776db08603f1a36f6d02071fa58d1d5b32", 16)
	pubkey := new(bn256.G1).ScalarMult(g1, secret)
	log.Println("pubkey:", pubkey.String())

	secret2, _ := new(big.Int).SetString("760a1080a8128d4b966bbe15243b9e776db08603f1a36f6d02071fa58d1d5882", 16)
	pubkey2 := new(bn256.G1).ScalarMult(g1, secret2)
	log.Println("pubkey2:", pubkey2.String())

	kadd := new(bn256.G1).Add(pubkey, pubkey2)
	log.Println("kadd:", kadd.String())
}