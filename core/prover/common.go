package prover

import (
	"encoding/hex"
	"github.com/hpb-project/HCash-SDK/core/types"
	"github.com/hpb-project/HCash-SDK/core/utils"
)

type TransferStatement struct {
	CLn   []types.Point
	CRn   []types.Point
	C     []types.Point
	D     types.Point
	Y     []types.Point
	Epoch uint
}

type TransferWitness struct {
	BTransfer int
	BDiff     int
	Index     []int
	SK        string // keypair['x'], bigInt hex string
	R         string // random scalar, bigInt hex string
}

type BurnWitness struct {
	SK    string // keypair['x'], bigInt hex string
	BDiff int
}

type BurnStatement struct {
	CLn    types.Point
	CRn    types.Point
	Y      types.Point
	Epoch  uint
	Sender string
}

var (
	b128 = utils.NewBN128()
)

func has0xPrefix(str string) bool {
	return len(str) >= 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X')
}

func FromHex(s string) []byte {
	if has0xPrefix(s) {
		s = s[2:]
	}
	if len(s)%2 == 1 {
		s = "0" + s
	}
	h, _ := hex.DecodeString(s)
	return h
}
