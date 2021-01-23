package core

import (
	"fmt"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/hpb-project/HCash-SDK/common"
	"github.com/hpb-project/HCash-SDK/common/types"
	"github.com/hpb-project/HCash-SDK/core/ebigint"
	"log"
)

type TransferStatement struct {
	CLn   []types.Point
	CRn   []types.Point
	C     []types.Point
	D     types.Point
	Y     []types.Point
	Epoch int
}

func (t *TransferStatement) Content() {
	log.Println("Transfer statement CLn = ")
	for i, cln := range t.CLn {
		fmt.Println("---> ", i, " = ", cln.String())
	}
	log.Println("Transfer statement CRn = ")
	for i, crn := range t.CRn {
		fmt.Println("---> ", i, " = ", crn.String())
	}
	log.Println("Transfer statement C = ")
	for i, c := range t.C {
		fmt.Println("---> ", i, " = ", c.String())
	}
	log.Println("Transfer statement Y = ")
	for i, y := range t.Y {
		fmt.Println("---> ", i, " = ", y.String())
	}
	log.Println("Transfer statement D = ", t.D.String())
	log.Println("Transfer statement epoch = ", t.Epoch)
}

type TransferWitness struct {
	BTransfer int
	BDiff     int
	Index     []int
	SK        string // keypair['x'], bigInt hex string
	R         string // random scalar, bigInt hex string
}

func (t *TransferWitness) Content() {
	log.Println("Transfer witness btransfer = ", t.BTransfer)
	log.Println("Transfer witness BDiff = ", t.BDiff)
	log.Println("Transfer witness Index = ", t.Index)
	log.Println("Transfer witness sk = ", t.SK)
	log.Println("Transfer witness R = ", t.R)
}

type BurnWitness struct {
	SK    string // keypair['x'], bigInt hex string
	BDiff int
}

type BurnStatement struct {
	CLn    types.Point
	CRn    types.Point
	Y      types.Point
	Epoch  int
	Sender string
}

var (
	b128 = NewBN128()
)

type ABI_Bytes32 [32]byte
type ABI_Bytes32_2 [2][32]byte
type ABI_Bytes32_2S [][2][32]byte
type ETH_ADDR ethcommon.Address

func parsePoints2ABI_Bytes32_2S(points []Point) []ABI_Bytes32_2 {
	var result = make([]ABI_Bytes32_2, len(points))
	{
		for i := 0; i < len(points); i++ {
			p := b128.Serialize(points[i])
			px := common.FromHex(p.GX())
			py := common.FromHex(p.GY())
			copy(result[i][0][:], px)
			copy(result[i][1][:], py)
		}
	}
	return result
}

func parsePoint2ABI_Bytes32_2(point Point) ABI_Bytes32_2 {
	var result ABI_Bytes32_2
	p := b128.Serialize(point)
	px := common.FromHex(p.GX())
	py := common.FromHex(p.GY())
	copy(result[0][:], px)
	copy(result[1][:], py)
	return result
}

func parseBigInt2ABI_Bytes32(e *ebigint.NBigInt) ABI_Bytes32 {
	var result ABI_Bytes32
	t := common.FromHex(b128.Bytes(e.Int))
	copy(result[:], t[:])
	return result
}
