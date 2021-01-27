package client

import (
	"encoding/json"
	"fmt"
	"github.com/hpb-project/HCash-SDK/core/ebigint"
	"gotest.tools/assert"
	"log"
	"testing"
)

func TestCreateAccount(t *testing.T) {

	for i := 0; i < 5; i++ {
		account := CreateAccount("")
		fmt.Println("create random account", account)
	}

	specialAccount := CreateAccount("0x299569ae0ae1d40140fd8d9afc54d2f581a292fd13fe88c7033d488119bb95b7")
	fmt.Println("specialAccount=", specialAccount)
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
	fmt.Println("sign result = ", result)
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

func TestTransferProof(t *testing.T) {
	var params = `{
		"epoch":53712840,
		"value":1,
		"diff" :3,
		"sk"   :"0x20a89bb465e9e2262e25901525509686f6a26b2fba976f1d9ff00a0cdbb362b0",
		"accounts": [
			[{"gx":"0x053225ab9382466d6b094e6e0ef738df4f3182757c6a0d48ab34f30691c422b5",
      		 "gy":"0x069427041ebfef2669c40b87a9d82690c75d332ecd4c1b775decd0e85884e2af"},
			 {"gx":"0x08d0fe696c3aff9c574949d3788e3d5379ee4d99ed9014b738f2825b8af231e7",
      		 "gy":"0x061aa65f5f632f30b3eef67207f2c3804f4e5ca07f4aaf195820b37c845fbd16"}
			],
			[{"gx":"0x0d9aa6c77eda65eee4299282135136c2b54205ccd835df94dc14bd0eb545553c",
      		 "gy":"0x2ead1bbaa97fa27f76e006f747a08cd8cdb45af4032cbfaffeeb6c1046f0d384"},
			 {"gx":"0x20bc85cf65b9afe7e4709592e382e59203e99e2c820c8bf2930707687004f687",
      		 "gy":"0x033e5bb2711dae6a5b6f25be4d150bdc43f505863aa1c8eced61ed74c051bee3"}
			]
		],
		"y":[
			{"gx":"0x2b621590db6b2e3ca3f0e562ed05487caa26ae88c6e1f54883a04e51f6664bc1",
			 "gy":"0x2c1173b211a55f5397ff869ae2feecad664a80730f4f6236a8664a167577ece7"},
			{"gx":"0x20710d65688c288d13a36884422807e5f49fb3785023d49067d1f1f1107cb484",
			 "gy":"0x09ad6933875e421a71f1ed619764ee73b0f628126ca9fe4c153368ed515e6db9"}
		],
		"index":[0, 1]
	}`
	b128.SetSpecialRandom(ebigint.FromHex("c3f4db6cd90e04d6e086f73fdb7a4ccaa4f57e48593d80c11c0fdd1fcac348df").ToRed(b128.Q()))
	result := TransferProof(params)
	log.Println("result = ", result)
}

func TestShuffle(t *testing.T) {
	var params = `{
		"self": {
			"gx":"0x06a50b18dca59140222133e77c3e3074e261caff764ef3520154a6650e7c2dc8",
			"gy":"0x01328650520364c24dfaeac3ed931220b1c425a7abf3154bb3ca254a0e0111a2"
		},
		"friend": {
			"gx":"0x0476db116e871903dac485ad7e2b14da1ef944e4ef113191e1ef2bb7036207a5",
			"gy":"0x1a173b29acf71a20d0393c8b78505295ba447ecea06928889931f025c10b1fec"
		},
		"decoys": [
			{"gx":"0x19447f92b16ab36ad31220030be6318903667005983915c026f852873eac3300",
			 "gy":"0x0bf4bdf1d4e2cdb6792253317c5dea7d479c42f619ba5f4e073be15432462c42"},
			{"gx":"0x184cea3a6b02dbe54da1e8818a94a5e3dfc093ad6cb84a9fb95e9ea8355de7df",
			 "gy":"0x23ece55b936c422d5a76204289777e63e9ac3be1db039fa8aa70bd282d28ea87"}
		]
	}`
	result := Shuffle(params)
	log.Println("shuffle result=", result)
}

func TestBurnProof(t *testing.T) {
	var params = `{
		"accounts":[
			{"gx":"0x19512743220081b7244cae299bb9f053b25d27337ee6b5d760eae272117db2af",
			 "gy":"0x0f7dae3691a53ec20f37e534c34c0eb41d256c8c7f9472e4c618126d6a054b58"},
			{"gx":"0x077da99d806abd13c9f15ece5398525119d11e11e9836b2ee7d23f6159ad87d4",
			  "gy":"0x01485efa927f2ad41bff567eec88f32fb0a0f706588b4e41a8d587d008b7f875"}
		],
		"epoch":53672920,
		"value":1,
		"diff":6,
		"sk":"0x04907c94209e3442e4830c142ba166ac032e511d00fcdf5f01b77d480518fa1a",
		"y":{
			"gx":"0x2af593d93442ca5d86d1f3748e624e68cc7db78da5fa568c40e32753e2e5b64b",
			"gy":"0x301248643b2813c1aaa9fbb7cec25fa6fb8e6d6db1240649b848a545962a9f81"},
		"sender":"d80ac1fb177c0b8d9c66de2b9657dd57084a2d7f"
	}`
	result := BurnProof(params)
	log.Println("burnproof result = ", result)
}
