package main

import "C"
import (
	"github.com/hpb-project/HCash-SDK/core/prover"
	"github.com/hpb-project/HCash-SDK/core/utils/service"
)

//export hCashBurnProof
func hCashBurnProof() string {
	s := prover.BurnProof()
	return s
}

//export hCashCreateAccount
func hCashCreateAccount(pwd string) string {

	var b_pwd = make([]byte, len(pwd))
	copy(b_pwd, []byte(pwd))

	account := service.CreateAccount(string(b_pwd))
	return account.String()
}

func main() {}
