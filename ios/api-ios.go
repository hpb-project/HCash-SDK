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
	return account
}

//export hCashSign
func hCashSign(input string) string {
	var data = make([]byte, len(input))
	copy(data, []byte(input))

	signed := client.Sign(string(data))
	return signed
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
	return result
}

//export hCashTransferProof
func hCashTransferProof(param string) string {
	var data = make([]byte, len(param))
	copy(data, []byte(param))

	result := client.TransferProof(string(data))
	return result
}

//export hCashBurnProof
func hCashBurnProof(param string) string {
	var data = make([]byte, len(param))
	copy(data, []byte(param))

	result := client.BurnProof(string(data))
	return result
}

//export hRegister
func hRegister(y string, c string, s string) string {
	result := client.Register(y, c, s)
	return result
}

//export hFund
func hFund(y string, b uint64) string {
	return client.Fund(y, b)
}

//export hTransfer
func hTransfer(c string, d string, y string, u string, proof string) string {
	return client.Transfer(c, d, y, u, proof)
}

//export hBurn
func hBurn(y string, bTransfer uint64, u string, proof string) string {
	return client.Burn(y, bTransfer, u, proof)
}

func main() {}
