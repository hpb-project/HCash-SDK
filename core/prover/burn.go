package prover

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/hpb-project/HCash-SDK/core/ebigint"
	"github.com/hpb-project/HCash-SDK/core/utils"
)

type BurnProof struct {
	BA       utils.Point
	BS       utils.Point
	tCommits *GeneratorVector

	tHat *ebigint.NBigInt
	mu   *ebigint.NBigInt

	c     *ebigint.NBigInt
	s_sk  *ebigint.NBigInt
	s_b   *ebigint.NBigInt
	s_tau *ebigint.NBigInt

	ipProof InnerProductProof
}

func (z BurnProof) Serialize() string {
	b128 := utils.NewBN128()
	result := "0x"
	result += b128.Representation(z.BA)[2:]
	result += b128.Representation(z.BS)[2:]

	tcv := z.tCommits.GetVector()
	for _, commit := range tcv {
		result += b128.Representation(commit)[2:]
	}

	result += b128.Bytes(z.tHat.Int)[2:]
	result += b128.Bytes(z.mu.Int)[2:]
	result += b128.Bytes(z.c.Int)[2:]
	result += b128.Bytes(z.s_sk.Int)[2:]
	result += b128.Bytes(z.s_b.Int)[2:]
	result += b128.Bytes(z.s_tau.Int)[2:]

	result += z.ipProof.Serialize()[2:]

	return result
}

type BurnProver struct {
	params   *GeneratorParams
	ipProver *InnerProductProver
}

func NewBurnProver() BurnProver {
	params := NewGeneratorParams(int(32), nil, nil)
	return BurnProver{
		params:   params,
		ipProver: new(InnerProductProver),
	}
}

func (b BurnProver) GenerateProof(statement map[string]interface{}, witness map[string]interface{}) {
	var proof = &BurnProof{}

	bytes32_2T, _ := abi.NewType("bytes32[2]", "", nil)
	uint256_T, _ := abi.NewType("uint256", "", nil)
	address_T, _ := abi.NewType("address", "", nil)

	arguments := abi.Arguments{
		{
			Type: bytes32_2T,
		},
		{
			Type: bytes32_2T,
		},
		{
			Type: bytes32_2T,
		},
		{
			Type: uint256_T,
		},
		{
			Type: address_T,
		},
	}
	vCLn := statement["CLn"].([2]string) //{{x1,y1}, {x2,y2}...}
	vCRn := statement["CRn"].([2]string)
	vy := statement["y"].([2]string)
	vepoch := statement["epoch"].(uint)
	vsender := statement["sender"].(string)

	bytes, _ := arguments.Pack(
		vCLn,
		vCRn,
		vy,
		vepoch,
		vsender)
	b128 := utils.NewBN128()
	var statementHash = utils.Hash(hex.EncodeToString(bytes))

}
