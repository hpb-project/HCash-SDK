package client

import (
	"encoding/json"
	"fmt"
	"gotest.tools/assert"
	"testing"
)

func TestCreateAccount(t *testing.T) {

	account := CreateAccount("")
	fmt.Println("create random account", account)

	specialAccount := CreateAccount("0x299569ae0ae1d40140fd8d9afc54d2f581a292fd13fe88c7033d488119bb95b7")

	type Accy struct {
		Gx string `json:"gx"`
		Gy string `json:"gy"`
	}
	type Acc struct {
		X string `json:"x"`
		Y Accy   `json"y"`
	}
	var acc Acc
	if err := json.Unmarshal([]byte(specialAccount), &acc); err != nil {
		t.Fatal("unmarshal accoutn failed, err :", err.Error())
	}
	assert.Equal(t, acc.X, "0x299569ae0ae1d40140fd8d9afc54d2f581a292fd13fe88c7033d488119bb95b7")
	assert.Equal(t, acc.Y.Gx, "0x042526b090bc34791599c53df82a129307914728eb9dcafe4a56d66d6c7cc76f")
	assert.Equal(t, acc.Y.Gy, "0x09c7fcbde6288f52f715f460f495714606b1e11897d1cff6fd80c576f6b9a896")
}

func TestReadBalance(t *testing.T) {
	var params = `{
		"CL": {
			"gx":"0x1b5d4b9abe488e61bbb92edff41682560a9d6e02335e2bca9b50881c9540e393",
			"gy":"0x15dc61a9eff5d5a4e70ed97cbce60f7afc69c9925a409ddba365897f1384ca58"
			},
		"CR": {
			"gx":"0x0456301d6013d1cc52455a37c8762f2463b1c7e148d55e1c7d9980d8ed8d54b8",
			"gy":"0x27e78199776a73737fa833429fd64e00fa592ca21dda2e92d3489c96148308cb"
			},
		"x":  "0x20a89bb465e9e2262e25901525509686f6a26b2fba976f1d9ff00a0cdbb362b0"
	}`
	balance := ReadBalance(params)
	assert.Equal(t, balance, 2)
}

func TestSign(t *testing.T) {
	var params = `{
		"address":"0xE4920905e06c6B6070477c40B85756ffDa3cD3E6",
		"account": {
			"x": "0x299569ae0ae1d40140fd8d9afc54d2f581a292fd13fe88c7033d488119bb95b7",
			"y": {
				"gx":"0x042526b090bc34791599c53df82a129307914728eb9dcafe4a56d66d6c7cc76f",
				"gy":"0x09c7fcbde6288f52f715f460f495714606b1e11897d1cff6fd80c576f6b9a896"
			}
		},
		"random":"0x2493a56987e869bbb150c14aff5b2e897d9fe78d6dad8b12c92432473f7e9abd"
	}`
	result := Sign(params)
	type SignResult struct {
		C string `json:"c"`
		S string `json:"s"`
	}
	var sr SignResult
	if err := json.Unmarshal([]byte(result), &sr); err != nil {
		t.Fatal("unmarshal sign result failed, err:", err)
	}
	assert.Equal(t, sr.C, "0x206db78bfe338ecffd5b2f0606789ff1045bfbf1e46c897f8fa2e2115e19ed74")
	assert.Equal(t, sr.S, "0x003fe7000561eeebccd4bff3160cd7f8fd50db62904d8fa217692a1f6ca8e7ed")
}
