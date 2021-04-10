package main

import "C"
import (
	"github.com/hpb-project/HCash-SDK/core/client"
)

//export hCashCreateAccount
func hCashCreateAccount(secret string) *C.char {
	var sk = make([]byte, len(secret))
	copy(sk, []byte(secret))

	account := client.CreateAccount(string(sk))
	return C.CString(account)
}

//export hCashSign
func hCashSign(input string) *C.char {
	var data = make([]byte, len(input))
	copy(data, []byte(input))

	signed := client.Sign(string(data))
	return C.CString(signed)
}

//export hCashReadBalance
func hCashReadBalance(param string) int32 {
	var data = make([]byte, len(param))
	copy(data, []byte(param))

	balance := client.ReadBalance(string(data))

	return int32(balance)
}

//export hCashShuffle
func hCashShuffle(param string) *C.char {
	var data = make([]byte, len(param))
	copy(data, []byte(param))

	result := client.Shuffle(string(data))
	return C.CString(result)
}

//export hCashTransferProof
func hCashTransferProof(param string) *C.char {
	var data = make([]byte, len(param))
	copy(data, []byte(param))

	result := client.TransferProof(string(data))
	return C.CString(result)
}

//export hCashBurnProof
func hCashBurnProof(param string) *C.char {
	var data = make([]byte, len(param))
	copy(data, []byte(param))

	result := client.BurnProof(string(data))
	return C.CString(result)
}

//export hRegister
func hRegister(y string, c string, s string) *C.char {
	return client.Register(y, c, s)
}

//export hFund
func hFund(y string, b uint64) *C.char {
	return client.Fund(y, b)
}

//export hTransfer
func hTransfer(c string, d string, y string, u string, proof string) *C.char {
	return client.Transfer(c, d, y, u, proof)
}

//export hBurn
func hBurn(y string, bTransfer uint64, u string, proof string) *C.char {
	return client.Burn(y, bTransfer, u, proof)
}

func main() {}
