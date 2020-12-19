package prover

import (
	"github.com/hpb-project/HCash-SDK/core/types"
	"github.com/hpb-project/HCash-SDK/core/utils"
)

type TransferStatement struct {
	CLn   []types.Publickey
	CRn   []types.Publickey
	C     []types.Publickey
	D     types.Publickey
	Y     []types.Publickey
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
	CLn    types.Publickey
	CRn    types.Publickey
	Y      types.Publickey
	Epoch  uint
	Sender string
}

var (
	b128 = utils.NewBN128()
)
