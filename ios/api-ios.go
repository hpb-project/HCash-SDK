package main

import "C"
import (
	"github.com/hpb-project/HCash-SDK/core/prover"
	"github.com/hpb-project/HCash-SDK/core/utils/service"
)

//export hCashBurnProof
func hCashBurnProof() *C.char {
	s := prover.BurnProof()
	return C.CString(s)
}

//export hCashCreateAccount
func hCashCreateAccount(pwd string) *C.char {

	var b_pwd = make([]byte,len(pwd))
	copy(b_pwd, []byte(pwd))

	account := service.CreateAccount(string(b_pwd))
	return C.CString(account.String())
}

func main() {}
