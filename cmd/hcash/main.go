package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	types2 "github.com/hpb-project/HCash-SDK/common/types"
	"github.com/hpb-project/HCash-SDK/core"
	"github.com/hpb-project/HCash-SDK/core/client"
	"log"
	"math/big"
	"strings"
	"time"
)

const (
	MainNet = "https://hpbnode.com"
)

var (
	SenderAddr  = common.Address{}
	ZSCContract = common.HexToAddress("0xe4920905e06c6b6070477c40b85756ffda3cd3e6")
)

func makeData(method string, data string) []byte {
	/*
		{
			"312a526c": "burn((bytes32,bytes32),uint256,(bytes32,bytes32),bytes)",
			"57d775f8": "epochLength()",
			"599c1a93": "fund((bytes32,bytes32),uint256)",
			"9b0d85d3": "register((bytes32,bytes32),uint256,uint256)",
			"79e543d0": "simulateAccounts((bytes32,bytes32)[],uint256)",
			"eff4d178": "transfer((bytes32,bytes32)[],(bytes32,bytes32),(bytes32,bytes32)[],(bytes32,bytes32),bytes)"
		}
	*/
	var input string
	switch method {
	case "burn":
		input = "312a526c"
	case "epochlength":
		input = "57d775f8"
	case "fund":
		input = "599c1a93"
	case "register":
		input = "9b0d85d3"
	case "simulateAccounts":
		input = "79e543d0"
	case "transfer":
		input = "eff4d178"
	default:
		return []byte{}
	}
	if len(data) > 2 && strings.HasPrefix(data, "0x") {
		input += data[2:]
	}
	return common.FromHex(input)
}

func CallSimulateAccounts(cli *HttpClient, y []types2.Point, epoch int64) ([][2]types2.Point, error) {
	param := &client.TxSimulateAccountsParam{
		Y:     y,
		Epoch: uint64(epoch),
	}
	str, _ := json.Marshal(param)

	res := client.TxSimulateAccounts(string(str))
	var saRes client.APIResponse
	if err := json.Unmarshal([]byte(res), &saRes); err != nil {
		log.Printf("txsimulateAccounts err %v\n", err)
		return nil, err
	}
	msg := ethereum.CallMsg{
		From:     SenderAddr,
		To:       &ZSCContract,
		Gas:      10000000,
		GasPrice: defaultgasprice,
		Value:    big.NewInt(0),
		Data:     makeData("simulateAccounts", saRes.Data),
	}
	simulates, err := cli.eth.CallContract(context.Background(), msg, nil)
	if err != nil {
		log.Printf("call contract failed, err = %v\n", err)
		return nil, err
	}

	parseParam := fmt.Sprintf("{\"data\":\"0x%s\"}", common.Bytes2Hex(simulates))
	parseRes := client.ParseSimulateAccountsData(parseParam)
	var paRes core.ParseSimulateAccountsResponse
	if err := json.Unmarshal([]byte(parseRes), &paRes); err != nil {
		log.Printf("parse simulateAccounts data failed, err %v\n", err)
		return nil, err
	}
	return paRes.Accounts, nil
}

func CallEpochLength(cli *HttpClient) (int64, error) {
	msg := ethereum.CallMsg{
		From:     SenderAddr,
		To:       &ZSCContract,
		Gas:      10000000,
		GasPrice: defaultgasprice,
		Value:    big.NewInt(0),
		Data:     makeData("epochlength", ""),
	}
	epochdata, err := cli.eth.CallContract(context.Background(), msg, nil)
	if err != nil {
		log.Println("call contract failed, err ", err)
		return 0, err
	}
	epoch := new(big.Int).SetBytes(epochdata).Int64()
	log.Println("get epoch = ", epoch)
	return epoch, nil
}

type HCashUser struct {
	Privk   string
	Balance int
	Y       types2.Point
	Epoch   int64
}

func (h *HCashUser) getEpoch() int64 {
	tm := time.Now().UnixNano() / 1000 / 1000
	return tm / 1000 / h.Epoch
}

func (h *HCashUser) burn(cli *HttpClient, value int, priv *ecdsa.PrivateKey) error {
	senderNonce := cli.GetNonce(SenderAddr.Hex())
	gasPrice := cli.GasPrice()
	chainId := cli.ChainID()

	var burnProofParam client.BurnProofParam
	var ep = h.getEpoch()
	burnProofParam.Y = h.Y
	burnProofParam.Epoch = int(ep)
	burnProofParam.Value = value
	burnProofParam.SK = h.Privk
	burnProofParam.Diff = h.Balance - value
	burnProofParam.Sender = SenderAddr.String()

	sim, err := CallSimulateAccounts(cli, []types2.Point{h.Y}, ep)
	burnProofParam.Accounts = sim[0][:]

	burnProofStr, _ := json.Marshal(burnProofParam)
	log.Printf("burn proof param = %v\n", burnProofParam)

	proofStr := client.BurnProof(string(burnProofStr))

	type Response struct {
		U     types2.Point `json:"u"`
		Proof string       `json:"proof"`
	}

	var proof Response
	if e := json.Unmarshal([]byte(proofStr), &proof); e != nil {
		log.Printf("unmarshal to BurnProofParam failed, err:%s\n", e.Error())
		return e
	}

	var txBurnParam client.TxBurnParam
	txBurnParam.B = uint64(value)
	txBurnParam.U = proof.U
	txBurnParam.Y = h.Y
	txBurnParam.Proof = proof.Proof
	paramdata, _ := json.Marshal(txBurnParam)

	txburnDataStr := client.TxBurn(string(paramdata))
	var txData client.APIResponse
	if e := json.Unmarshal([]byte(txburnDataStr), &txData); e != nil {
		log.Printf("unmarshal to BurnProofParam failed, err:%s\n", e.Error())
		return err
	}

	tx := types.NewTransaction(senderNonce, ZSCContract, big.NewInt(0), 50000000, gasPrice,
		makeData("burn", txData.Data))

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainId), priv)
	if err != nil {
		log.Printf("sign tx failed, err = %v\n", err)
		return err
	}
	txhash, err := cli.SendSignedTx(signedTx)
	if err != nil {
		log.Printf("send signed tx failed, err = %v\n", err)
		return err
	}
	log.Printf("send burn tx with txhash(%s)\n", txhash)

	return nil
}

func main() {

	senderPrivKey := flag.String("sk", "", "Sender private key in hex")
	alicePrivKey := flag.String("ak", "", "alice private key in hex")

	flag.Parse()

	senderPriv, err := crypto.HexToECDSA(*senderPrivKey)
	if err != nil {
		log.Printf("invalid send privatekey")
		return
	}
	SenderAddr = getAddrFromPrivkey(senderPriv)

	cli := NewHttpClient(MainNet)
	alice, err := RecoverUser(*alicePrivKey)
	if err != nil {
		log.Printf("recover user failed, err = %v\n", err)
		return
	}

	alice.Epoch, err = CallEpochLength(cli)
	if err != nil {
		log.Printf("get epoch failed, err %v\n", err)
		return
	}

	sim, err := CallSimulateAccounts(cli, []types2.Point{alice.Y}, alice.getEpoch()+1)
	if err != nil {
		log.Printf("get cl failed err = %v\n", err)
		return
	}
	log.Printf("got simulate accounts %v\n", sim)

	alice.Balance = ReadBalance(sim[0][0], sim[0][1], alice.Privk)
	log.Println("got alice.Balance = ", alice.Balance)

	// test burn
	if alice.Balance > 1 {
		err := alice.burn(cli, 1, senderPriv)
		if err != nil {
			log.Println("alice burn failed, err ", err)
			return
		}
	}
}

func ReadBalance(cl, cr types2.Point, x string) int {
	var readBalance client.ReadBalanceParam
	readBalance.X = x
	readBalance.CL = cl
	readBalance.CR = cr
	param, _ := json.Marshal(readBalance)
	return client.ReadBalance(string(param))
}

func RecoverUser(privk string) (*HCashUser, error) {
	var acc core.Account
	s := client.CreateAccount(privk)
	if err := json.Unmarshal([]byte(s), &acc); err != nil {
		log.Println("create Account failed, err ", err)
		return nil, err
	}
	user := &HCashUser{
		Y:     acc.Y,
		Privk: privk,
		Epoch: 20,
	}
	return user, nil
}

func getAddrFromPrivkey(priv *ecdsa.PrivateKey) common.Address {
	publicKey := priv.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	fmt.Println("from address", fromAddress.String())
	return fromAddress
}
