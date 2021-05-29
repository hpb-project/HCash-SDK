package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum/core/types"

	//	"errors"
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
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
	MainNet = "http://114.242.26.15:30180"
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
	Friends map[string]types2.Point
}

func (h *HCashUser) addFriend(name string, xy []string) {
	if _, exist := h.Friends[name]; exist {
		return
	}
	var y types2.Point
	y.Set(xy)
	h.Friends[name] = y
}

func (h *HCashUser) getEpoch() int64 {
	tm := time.Now().UnixNano() / 1000 / 1000 / 1000
	fmt.Printf("current tmstamp %v\n", tm)
	return tm / h.Epoch
}

func sendTx(cli *HttpClient, method string, priv *ecdsa.PrivateKey, data string) error {
	senderNonce := cli.GetNonce(SenderAddr.Hex())
	gasPrice := cli.GasPrice()
	chainId := cli.ChainID()

	txdata := makeData(method, data)
	fmt.Println("txdata = ", hex.EncodeToString(txdata))

	tx := types.NewTransaction(senderNonce, ZSCContract, big.NewInt(0), 50000000, gasPrice,
		txdata)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainId), priv)
	if err != nil {
		log.Printf("sign tx failed, err = %v\n", err)
		return err
	}
	fmt.Println("signTx = ", signedTx.Hash())
	//time.Sleep(3*time.Second )
	txhash, err := cli.SendSignedTx(signedTx)
	if err != nil {
		log.Printf("send signed tx failed, err = %v\n", err)
		return err
	}
	log.Printf("send burn tx with txhash(%s)\n", txhash)
	return nil
}

func (h *HCashUser) transfer(cli *HttpClient, value int, friend string, priv *ecdsa.PrivateKey) error {
	if f, exist := h.Friends[friend]; !exist {
		return errors.New(fmt.Sprintf("not found friend %s", friend))
	} else {
		var shuffleParam client.ShuffleParam
		shuffleParam.Self = h.Y
		shuffleParam.Friend = f
		shuffleParam.Decoys = []types2.Point{}

		sstr, _ := json.Marshal(shuffleParam)
		sres := client.Shuffle(string(sstr))

		type response struct {
			Y     []types2.Point `json:"y"`
			Index []int          `json:"index"`
		}
		var shuffleRes response
		json.Unmarshal([]byte(sres), &shuffleRes)
		fmt.Printf("shuffled = %v\n", shuffleRes)

		var ep = h.getEpoch() // int64(54076096)
		sims, err := CallSimulateAccounts(cli, shuffleRes.Y, ep)
		if err != nil {
			fmt.Printf("callSimulateAccounts failed, err = %v\n", err.Error())
			return err
		}
		fmt.Printf("simulates = %v\n", sims)

		var transferProofParam client.TransferProofParam
		transferProofParam.SK = h.Privk
		transferProofParam.Value = value
		transferProofParam.Diff = h.Balance - value
		transferProofParam.Epoch = int(ep)
		transferProofParam.Accounts = sims
		transferProofParam.Y = shuffleRes.Y
		transferProofParam.Index = shuffleRes.Index

		trpstr, _ := json.Marshal(transferProofParam)
		trpresStr := client.TransferProof(string(trpstr))

		fmt.Printf("transfer proof = %v\n", trpresStr)

		type Response struct {
			C     []types2.Point `json:"C"`
			D     types2.Point   `json:"D"`
			U     types2.Point   `json:"u"`
			Y     []types2.Point `json:"y"`
			Proof string         `json:"proof"`
		}
		var trpRes Response
		json.Unmarshal([]byte(trpresStr), &trpRes)

		var txTransferParam client.TxTransferParam
		txTransferParam.U = trpRes.U
		txTransferParam.Y = trpRes.Y
		txTransferParam.Proof = trpRes.Proof
		txTransferParam.C = trpRes.C
		txTransferParam.D = trpRes.D

		txpstr, _ := json.Marshal(txTransferParam)

		txdataStr := client.TxTransfer(string(txpstr))
		fmt.Printf("txtransfer data = %v\n", txdataStr)

		var txData client.APIResponse
		if e := json.Unmarshal([]byte(txdataStr), &txData); e != nil {
			log.Printf("unmarshal to BurnProofParam failed, err:%s\n", e.Error())
			return err
		}
		return sendTx(cli, "transfer", priv, txData.Data)
	}
}

func (h *HCashUser) burn(cli *HttpClient, value int, priv *ecdsa.PrivateKey) error {
	var burnProofParam client.BurnProofParam
	var ep = h.getEpoch() //54073887 //
	fmt.Println("epoch = ", ep)
	burnProofParam.Y = h.Y
	burnProofParam.Epoch = int(ep)
	burnProofParam.Value = value
	burnProofParam.SK = h.Privk
	burnProofParam.Diff = h.Balance - value
	burnProofParam.Sender = SenderAddr.String() //"0x38462d46fc145fc71e85643cd1efb9b0c61e5ed0"//SenderAddr.String()

	sim, err := CallSimulateAccounts(cli, []types2.Point{h.Y}, int64(ep))
	burnProofParam.Accounts = sim[0][:]

	burnProofStr, _ := json.Marshal(burnProofParam)
	//log.Printf("burn proof param = %v\n", burnProofParam)

	proofStr := client.BurnProof(string(burnProofStr))
	//log.Printf("get burnProof = %v\n", proofStr)

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
	sendTx(cli, "burn", priv, txData.Data)
	//fmt.Println("txburnDataStr ", txburnDataStr)

	return nil
}

func main() {

	senderPrivKey := flag.String("sk", "", "Sender private key in hex")
	alicePrivKey := flag.String("ak", "", "alice private key in hex")
	doBurn := flag.Bool("b", false, "do burn if balance > 0")
	doTx := flag.Bool("t", false, "do transfer if balance > 0")

	flag.Parse()

	if strings.HasPrefix(*senderPrivKey, "0x") ||
		strings.HasPrefix(*senderPrivKey, "0X") {
		*senderPrivKey = (*senderPrivKey)[2:]
	}

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

	alice.addFriend("bob", []string{
		"0x28d55db8435a8fdd93bf0e40d339d2f1cd8a7033ef47e1635695a94add9b85a5",
		"0x23f58460eda5eb93a8995649ef25d51221aa206b6242080322eea2cd910019d2",
	})

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
	if alice.Balance > 0 && *doBurn {
		err := alice.burn(cli, 1, senderPriv)
		if err != nil {
			log.Println("alice burn failed, err ", err)
			return
		}
	}

	if alice.Balance > 0 && *doTx {
		err := alice.transfer(cli, 1, "bob", senderPriv)
		if err != nil {
			log.Println("alice transfer failed, err ", err)
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
		Y:       acc.Y,
		Privk:   privk,
		Epoch:   20,
		Friends: make(map[string]types2.Point),
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
