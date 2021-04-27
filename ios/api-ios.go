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

//export hCashTxRegister
func hCashTxRegister(param string) string {
	var data = make([]byte, len(param))
	copy(data, []byte(param))

	result := client.TxRegister(string(data))
	return result
}

//export hCashTxFund
func hCashTxFund(param string) string {
	var data = make([]byte, len(param))
	copy(data, []byte(param))

	result := client.TxFund(string(data))
	return result
}

//export hCashTxTransfer
func hCashTxTransfer(param string) string {
	var data = make([]byte, len(param))
	copy(data, []byte(param))

	result := client.TxTransfer(string(data))
	return result
}

//export hCashTxBurn
func hCashTxBurn(param string) string {
	var data = make([]byte, len(param))
	copy(data, []byte(param))

	result := client.TxBurn(string(data))
	return result
}
//export hCashTxSimulateAccounts
func hCashTxSimulateAccounts(param string) string {
	var data = make([]byte, len(param))
	copy(data, []byte(param))

	result := client.TxSimulateAccounts(string(data))
	return result
}

func main() {}
