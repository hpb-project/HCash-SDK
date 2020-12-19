package utils

import "github.com/hpb-project/HCash-SDK/core/prover"

func ProveTransfer(statement prover.TransferStatement, witness prover.TransferWitness) string {
	zether := prover.NewZetherProver()
	proof := zether.GenerateProof(statement, witness)
	if proof != nil {
		return ""
	}
	return proof.Serialize()
}

func ProveBurn(statement prover.BurnStatement, witness prover.BurnWitness) string {
	burn := prover.NewBurnProver()
	proof := burn.GenerateProof(statement, witness)
	if proof != nil {
		return ""
	}
	return proof.Serialize()
}
