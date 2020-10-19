package service

import (
	"encoding/hex"
	"encoding/json"
)
type Key []byte

type Account struct {
	Pubkey		Key		`json:"pubkey"`
	PrivateKey 	Key		`json:"privatekey"`
}
func (k Key) String() string {
	ks := hex.EncodeToString(k[:])
	return ks
}

func (a Account) String() string {
	b,_ := json.Marshal(a)
	return string(b)
}

func CreateAccount(pwd string) *Account {
	pubk,_ := hex.DecodeString("77da99d806abd13c9f15ece5398525119d11e11e9836b2ee7d23f6159ad87d4")
	skey,_ := hex.DecodeString("1485efa927f2ad41bff567eec88f32fb0a0f706588b4e41a8d587d008b7f875")

	return &Account{
		Pubkey: pubk,
		PrivateKey: skey,
	}
}
