package core

import (
	"github.com/hpb-project/HCash-SDK/common"
)

var (
	BASEPOS = 64
)

func Register(y string, c string, s string) string {
	if common.Has0xPrefix(y) {
		y = y[2:]
	}
	if common.Has0xPrefix(c) {
		c = c[2:]
	}
	if common.Has0xPrefix(s) {
		s = s[2:]
	}
	result := y + c + s
	return result
}

func Fund(y string, b uint64) string {
	if common.Has0xPrefix(y) {
		y = y[2:]
	}
	btransfer := common.Uint642Bytes32(b)
	result := y + btransfer
	return result
}

func Transfer(c string, d string, y string, u string, proof string) string {
	if common.Has0xPrefix(c) {
		c = c[2:]
	}
	if common.Has0xPrefix(d) {
		d = d[2:]
	}
	if common.Has0xPrefix(y) {
		y = y[2:]
	}
	if common.Has0xPrefix(u) {
		u = u[2:]
	}
	if common.Has0xPrefix(proof) {
		proof = proof[2:]
	}
	cpos := BASEPOS + len(d) + BASEPOS + len(u) + BASEPOS
	ypos := cpos + BASEPOS + len(c)
	proofpos := ypos + BASEPOS + len(y)
	result := common.Uint642Bytes32(uint64(cpos)/2) + d
	result = result + common.Uint642Bytes32(uint64(ypos)/2) + u
	result = result + common.Uint642Bytes32(uint64(proofpos)/2)
	result = result + common.Uint642Bytes32(uint64(len(c)/64)/2) + c
	result = result + common.Uint642Bytes32(uint64(len(y)/64)/2) + y
	result = result + common.Uint642Bytes32(uint64(len(proof))/2) + proof
	return result
}
func Burn(y string, bTransfer uint64, u string, proof string) string {
	if common.Has0xPrefix(y) {
		y = y[2:]
	}
	if common.Has0xPrefix(u) {
		u = u[2:]
	}
	if common.Has0xPrefix(proof) {
		proof = proof[2:]
	}

	result := y + common.Uint642Bytes32(bTransfer)
	result = result + u
	result = result + common.Uint642Bytes32(uint64(len(result)+BASEPOS)/2)
	result = result + common.Uint642Bytes32(uint64(len(proof))/2)
	result = result + proof
	return result
}

func SimulateAccounts(y string, epoch uint64) string {
	if common.Has0xPrefix(y) {
		y = y[2:]
	}
	// 79e543d0 ypos epoch len(y) y...
	ypos := BASEPOS
	ylength := len(y) / 64 / 2
	result := common.Uint642Bytes32(uint64(ypos))
	result = result + common.Uint642Bytes32(epoch)
	result = result + common.Uint642Bytes32(uint64(ylength))
	result = result + y
	return result
}
