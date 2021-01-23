package core

func ProveTransfer(statement TransferStatement, witness TransferWitness) string {
	zether := NewZetherProver()
	//statement.Content()
	//witness.Content()
	proof := zether.GenerateProof(statement, witness)
	if proof == nil {
		return ""
	}
	return proof.Serialize()
}

func ProveBurn(statement BurnStatement, witness BurnWitness) string {
	burn := NewBurnProver()
	proof := burn.GenerateProof(statement, witness)
	if proof == nil {
		return ""
	}
	return proof.Serialize()
}
