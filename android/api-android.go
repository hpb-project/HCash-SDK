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

//export hCashTxRegister
func hCashTxRegister(param string) *C.char {
	var data = make([]byte, len(param))
	copy(data, []byte(param))

	result := client.TxRegister(string(data))
	return C.CString(result)
}

//export hCashTxFund
func hCashTxFund(param string) *C.char {
	var data = make([]byte, len(param))
	copy(data, []byte(param))

	result := client.TxFund(string(data))
	return C.CString(result)
}

//export hCashTxTransfer
func hCashTxTransfer(param string) *C.char {
	var data = make([]byte, len(param))
	copy(data, []byte(param))

	result := client.TxTransfer(string(data))
	return C.CString(result)
}

//export hCashTxBurn
func hCashTxBurn(param string) *C.char {
	var data = make([]byte, len(param))
	copy(data, []byte(param))

	result := client.TxBurn(string(data))
	return C.CString(result)
}

//export hCashTxSimulateAccounts
func hCashTxSimulateAccounts(param string) *C.char {
	var data = make([]byte, len(param))
	copy(data, []byte(param))

	result := client.TxSimulateAccounts(string(data))
	return C.CString(result)
}

//export hCashParseSimulateAccountsData
func hCashParseSimulateAccountsData(param string) *C.char {
	var data = make([]byte, len(param))
	copy(data, []byte(param))

	result := client.ParseSimulateAccountsData(string(data))
	return C.CString(result)
}

func main() {}
