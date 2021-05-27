package core

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/hpb-project/HCash-SDK/common"
	"github.com/hpb-project/HCash-SDK/common/types"
	"math/big"
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

type ParseSimulateAccountsResponse struct {
	Accounts [][2]types.Point `json:"accounts"`
}

func ParseSimulateAccounts(data string) (*ParseSimulateAccountsResponse, error) {
	hexdata := common.FromHex(data)
	fmt.Println("params = ", hex.EncodeToString(hexdata))
	offset := new(big.Int).SetBytes(hexdata[:32]).Int64()
	if offset > int64(len(hexdata)) {
		return nil, errors.New(fmt.Sprintf("invalid param %s", data))
	}
	hexdata = hexdata[offset:]

	if len(hexdata) < 32 {
		return nil, errors.New(fmt.Sprintf("invalid param %s", data))
	}
	length := new(big.Int).SetBytes(hexdata[:32]).Int64()
	hexdata = hexdata[32:]
	fmt.Printf("offset = %d, length = %d\n", offset, length)

	if length*128 != int64(len(hexdata)) {
		return nil, errors.New(fmt.Sprintf("invalid param %s", data))
	}
	res := &ParseSimulateAccountsResponse{
		Accounts: make([][2]types.Point, 0),
	}

	start := hexdata
	for len(start) > 0 {
		item := [2]types.Point{}
		account0_x := "0x" + hex.EncodeToString(start[:32])
		account0_y := "0x" + hex.EncodeToString(start[32:64])
		item[0].Set([]string{account0_x, account0_y})
		//fmt.Printf("a0_x = %s, a0_y = %s, item = %s\n", account0_x, account0_y, item[0])

		account1_x := "0x" + hex.EncodeToString(start[64:96])
		account1_y := "0x" + hex.EncodeToString(start[96:128])
		item[1].Set([]string{account1_x, account1_y})
		//fmt.Printf("a1_x = %s, a1_y = %s, item = %s\n", account1_x, account1_y, item[1])
		res.Accounts = append(res.Accounts, item)
		start = start[128:]
	}
	return res, nil

}
