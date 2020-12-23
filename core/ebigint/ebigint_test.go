package ebigint

import (
	"github.com/hpb-project/HCash-SDK/common"
	"log"
	"testing"
)

func TestBase(t *testing.T) {
	var secret = "100a1080a8128d4b966bbe15243b9e776db08603f1a36f6d02071fa58d1d5b32"
	q := FromBytes(common.FromHex("30644e72e131a029b85045b68181585d2833e84879b9709143e1f593f0000001"))
	x := FromBytes(common.FromHex(secret)).ToRed(q.Red(q.Int))
	hexstr := x.Text(16)
	log.Println("x = ", hexstr)
	if hexstr != "100a1080a8128d4b966bbe15243b9e776db08603f1a36f6d02071fa58d1d5b32" {
		t.Error("ToRed is not match")
	}
}
