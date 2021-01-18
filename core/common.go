package core

import (
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/hpb-project/HCash-SDK/common"
	"github.com/hpb-project/HCash-SDK/common/types"
	"github.com/hpb-project/HCash-SDK/core/ebigint"
)

type TransferStatement struct {
	CLn   []types.Point
	CRn   []types.Point
	C     []types.Point
	D     types.Point
	Y     []types.Point
	Epoch int
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
