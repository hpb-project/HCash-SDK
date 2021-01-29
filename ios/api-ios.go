package main

import "C"
import (
	"github.com/hpb-project/HCash-SDK/core/client"
)

//export hCashCreateAccount
func hCashCreateAccount(secret string) string {
	var sk = make([]byte, len(secret))
	copy(sk, []byte(secret))

	account := client.CreateAccount(string(sk))
	return C.CString(account)
}

//export hCashSign
func hCashSign(input string) string {
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
func hCashShuffle(param string) string {
	var data = make([]byte, len(param))
	copy(data, []byte(param))

	result := client.Shuffle(string(data))
	return C.CString(result)
}

//export hCashTransferProof
func hCashTransferProof(param string) string {
	var data = make([]byte, len(param))
	copy(data, []byte(param))

	result := client.TransferProof(string(data))
	return C.CString(result)
}

//export hCashBurnProof
func hCashBurnProof(param string) string {
	var data = make([]byte, len(param))
	copy(data, []byte(param))

	result := client.BurnProof(string(data))
	return C.CString(result)
}

func main() {}
