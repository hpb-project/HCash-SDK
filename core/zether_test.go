package core

import (
	"github.com/hpb-project/HCash-SDK/common/types"
	"log"
	"testing"
)

func TestZetherProof(t *testing.T) {
	zeth := NewZetherProver()
	var CLn = make([]types.Point, 2)
	CLn[0] = types.Point{"0x2b6dc01a49982bfcbfb49a091a80758244ea78ee166931c4d679a7d2681fcccf", "0x0278ef49a7bbf8ccd4003ec6cd4689595062811c39f68664a1a8dc6d11447933"}
	CLn[1] = types.Point{"0x0dd30ebd35990f92ff8e398908635d1bd949b77663f0a060ef2872ca965f1ffb", "0x00246f9105a20fa6fe289a6812e0a8885127ed0c3b6a99735bc08c7ceb58cf59"}
	var CRn = make([]types.Point, 2)
	CRn[0] = types.Point{"0x14fe37158cc51254aa0ed3ca5b228bdf21e8d27f804c90d1a71dac78ce40f1b5", "0x00ab2fb4e6ccc59851dee900c4d7b8b080c779217ee3704bd373b1aedc76fd3e"}
	CRn[1] = types.Point{"0x14fe37158cc51254aa0ed3ca5b228bdf21e8d27f804c90d1a71dac78ce40f1b5", "0x00ab2fb4e6ccc59851dee900c4d7b8b080c779217ee3704bd373b1aedc76fd3e"}
	var C = make([]types.Point, 2)
	C[0] = types.Point{"0x0c59ba24b1ef2f85534cbf11017cc3a41f3987ea6b04f6fa0dcedc416dbd5624", "0x2941614917c49efcc09db3d74654b04ce864cb9b43de677db2172cfed3e73efd"}
	C[1] = types.Point{"0x2115e60097e2f075e227829b83b52d4f691b79fb4b530a23bd5a5eb655b3445b", "0x2eb55ed5709bedba6307d59392bf55b992579c1e439d0f775450822baccfa52f"}
	var D = types.Point{"0x0b5f411bf6c261a2b865cea00230d5a0c1fedcb6caee8c4c54e818d09dd8f8c3", "0x06ddffef614e0e1ca581057871a6f067470fe09f7a847c4e6475b1367a01a6f9"}
	var y = make([]types.Point, 2)
	y[0] = types.Point{"0x07121c805d96cbf8204eec59ebc495d2b9dfb365c6521af2609fd172bd2c2887", "0x24ee9f1862c1bd9ed0ad88b7f90376e3f237637f47ddc03a4b4353355d415272"}
	y[1] = types.Point{"0x2b621590db6b2e3ca3f0e562ed05487caa26ae88c6e1f54883a04e51f6664bc1", "0x2c1173b211a55f5397ff869ae2feecad664a80730f4f6236a8664a167577ece7"}
	var epoch = 53687137

	var sk = "20a89bb465e9e2262e25901525509686f6a26b2fba976f1d9ff00a0cdbb362b0"
	var r = "2296c63311038849058a5a831a333ee2c3643aba005bf3cf98dd4c6972a79d1d"

	var bTransfer = 1
	var bDiff = 1
	var index = []int{1, 0}

	var istatement = TransferStatement{
		CLn:   CLn,
		CRn:   CRn,
		C:     C,
		D:     D,
		Y:     y,
		Epoch: epoch,
	}
	var iwitness = TransferWitness{
		BTransfer: bTransfer,
		BDiff:     bDiff,
		Index:     index,
		SK:        sk,
		R:         r,
	}
	proof := zeth.GenerateProof(istatement, iwitness)
	if proof != nil {
		ss := proof.Serialize()
		log.Println("proof = ", ss)
	} else {
		log.Println("generate proof failed.")
	}
}
