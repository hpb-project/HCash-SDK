package core

import (
	"testing"
)

func TestBase(t *testing.T) {
	b128 := NewBN128()
	q := b128.Q()
	if q.Text(16) != "30644e72e131a029b85045b68181585d2833e84879b9709143e1f593f0000001" {
		t.Error("b128.q is not match")
	}

	//var secret = "100a1080a8128d4b966bbe15243b9e776db08603f1a36f6d02071fa58d1d5b32"
	//x := ebigint.FromBytes(common.FromHex(secret)).ToRed(b128.Q())
	//hexstr := x.Text(16)
	//log.Println("x = ", hexstr)
	//if hexstr != "100a1080a8128d4b966bbe15243b9e776db08603f1a36f6d02071fa58d1d5b32" {
	//	t.Error("ToRed is not match")
	//}
}
