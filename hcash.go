package main

import (
	"github.com/hpb-project/HCash-SDK/common/types"
	"github.com/hpb-project/HCash-SDK/core"
)

func main() {
	cln := types.Point{"0x1418a69e20ab642d7dad6e8080de42a0f6a2110dcb20e35bda8e3a9a47161f26", "0x0207f80673298caa563db3537892881a13d2b234c5e7ab1b5e368ff072542558"}
	crn := types.Point{"0x077da99d806abd13c9f15ece5398525119d11e11e9836b2ee7d23f6159ad87d4", "0x01485efa927f2ad41bff567eec88f32fb0a0f706588b4e41a8d587d008b7f875"}
	y := types.Point{"0x2af593d93442ca5d86d1f3748e624e68cc7db78da5fa568c40e32753e2e5b64b", "0x301248643b2813c1aaa9fbb7cec25fa6fb8e6d6db1240649b848a545962a9f81"}
	epoch := 53672920
	home := "d80ac1fb177c0b8d9c66de2b9657dd57084a2d7f"
	x := "0x04907c94209e3442e4830c142ba166ac032e511d00fcdf5f01b77d480518fa1a"
	diff := 6

	istatement := core.BurnStatement{cln, crn, y, epoch, home}

	iwitness := core.BurnWitness{x, diff}
	proof := core.NewBurnProver()

	proof.GenerateProof(istatement, iwitness)

}
