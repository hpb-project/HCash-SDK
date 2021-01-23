package client

import (
	"encoding/json"
	"github.com/hpb-project/HCash-SDK/common"
	"github.com/hpb-project/HCash-SDK/common/types"
	"github.com/hpb-project/HCash-SDK/core"
	"github.com/hpb-project/HCash-SDK/core/ebigint"
	"log"
	"math"
	"math/big"
	"math/rand"
)

var (
	b128 = core.NewBN128()
)

/*
 * input : a big number hex string or empty string.
 * output: {'x':'', 'y':{'gx':'', 'gy':''}}
 */
func CreateAccount(secret string) string {
	var account core.Account
	if secret != "" {
		x, _ := new(big.Int).SetString(common.HexWithout0x(secret), 16)
		account.X = ebigint.ToNBigInt(x).ToRed(b128.Q())
		account.Y = b128.Serialize(b128.CurveG().Mul(account.X))
	} else {
		account = core.CreateAccount()
	}

	data, _ := json.Marshal(account)
	return string(data)
}

/*
 * input:
	zscAddress : zsc contract address string,
	account    : account json string. {'x':'', 'y': {'gx':'',  'gy':''}}
 * output:
	json string, content is big number hex string. {'c':'', 's':''}
*/
type SignParam struct {
	ZSCAddr   string       `json:"address"`
	Accounter core.Account `json:"account"`
	Random    string       `json:"random"`
}

func Sign(input string) string {
	var param SignParam
	var c, s *ebigint.NBigInt
	var e error
	if e = json.Unmarshal([]byte(input), &param); e != nil {
		log.Printf("unmarshal param failed, err:%s\n", e.Error())
		return ""
	}
	if param.Random != "" {
		nk, ok := new(big.Int).SetString(common.HexWithout0x(param.Random), 16)
		if !ok {
			c, s, e = core.Sign(common.FromHex(param.ZSCAddr), param.Accounter)
		} else {
			sign_k := ebigint.ToNBigInt(nk)
			c, s, e = core.SignWithRandom(common.FromHex(param.ZSCAddr), param.Accounter, sign_k)
		}
	} else {
		c, s, e = core.Sign(common.FromHex(param.ZSCAddr), param.Accounter)
	}

	if e != nil {
		log.Println("sign failed error:", e.Error())
		return ""
	}

	type CS struct {
		C string `json:"c"`
		S string `json:"s"`
	}
	var ret_cs = CS{
		C: b128.Bytes(c.Int),
		S: b128.Bytes(s.Int),
	}
	data, _ := json.Marshal(ret_cs)
	return string(data)
}

/*
 * input: param is json string, {''}
 */
type ReadBalanceParam struct {
	CL types.Point `json:"CL"`
	CR types.Point `json:"CR"`
	X  string      `json:"x"`
}

func ReadBalance(param string) int {
	var p ReadBalanceParam
	if e := json.Unmarshal([]byte(param), &p); e != nil {
		log.Printf("unmarshal param failed, err:%s\n", e.Error())
		return 0
	}
	x := ebigint.FromHex(p.X).ForceRed(b128.Q())

	return core.ReadBalance(p.CL, p.CR, x)
}

/*
 * input: {'self':{'gx':'', 'gy':''},
			'friend':{'gx':'', 'gy':''},
			'decoys':[
					{'gx':'', 'gy':''},
					{'gx':'', 'gy':''}]}
	output: { 'y': [
					{'gx':'', 'gy':''},
					{'gx':'', 'gy':''},....],
			  'index':[10, 20]}
*/
type ShuffleParam struct {
	Self   types.Point   `json:"self"`
	Friend types.Point   `json:"friend"`
	Decoys []types.Point `json:"decoys"`
}

func Shuffle(param string) string {
	var p ShuffleParam
	if e := json.Unmarshal([]byte(param), &p); e != nil {
		return ""
	}
	var y = make([]types.Point, 0)
	y = append(y, p.Self)
	y = append(y, p.Friend)
	for _, decoy := range p.Decoys {
		y = append(y, decoy)
	}

	var index [2]int
	var m = len(y)

	for m != 0 {
		// https://bost.ocks.org/mike/shuffle/
		var i = int(math.Floor(rand.Float64() * float64(m)))
		m -= 1

		var temp = y[i]
		y[i] = y[m]
		y[m] = temp

		if temp.Match(p.Self) {
			index[0] = m
		} else if temp.Match(p.Friend) {
			index[1] = m
		}
	} // shuffle the array of y's
	if (index[0] % 2) == (index[1] % 2) {
		var temp = y[index[1]]
		var delta = 0
		if index[1]%2 == 0 {
			delta = 1
		} else {
			delta = -1
		}

		y[index[1]] = y[index[1]+delta]
		y[index[1]+delta] = temp
		index[1] = index[1] + delta
	} // make sure you and your friend have opposite parity

	type response struct {
		Y     []types.Point `json:"y"`
		Index []int         `json:"index"`
	}
	var res response
	res.Y = y
	res.Index = make([]int, len(index))
	for i := 0; i < len(index); i++ {
		res.Index[i] = index[i]
	}
	b, _ := json.Marshal(res)
	return string(b)
}

type TransferProofParam struct {
	Epoch    int              `json:"epoch"`
	Value    int              `json:"value"`
	Diff     int              `json:"diff"`
	SK       string           `json:"sk"`
	Y        []types.Point    `json:"y"`
	Index    []int            `json:"index"`
	Accounts [][2]types.Point `json:"accounts"`
}

func TransferProof(param string) string {
	var p TransferProofParam
	if e := json.Unmarshal([]byte(param), &p); e != nil {
		log.Printf("unmarshal param to TransferProofParam failed, err:%s\n", e.Error())
		return ""
	}
	var unserialized = make([][2]core.Point, 0)
	for i, account := range p.Accounts {
		var m [2]core.Point
		m[0] = b128.UnSerialize(account[0])
		m[1] = b128.UnSerialize(account[1])
		unserialized = append(unserialized, m)
		log.Println("unserialized ---> ", i, " account[0] = ", account[0], " account[1] = ", account[1])
	}
	if Some(unserialized) {
		log.Printf("Reject, please make sure all parties(include decoys) are registered\n")
		return ""
	}

	var r = b128.RandomScalar()
	log.Println("r = ", r.Text(16))

	var C = make([]core.Point, len(p.Y))
	for i, party := range p.Y {
		//var C = y.map((party, i) => bn128.curve.g.mul(
		//	i == index[0] ? new BN(-value) :
		//		i == index[1] ? new BN(value) : new BN(0)
		//	).add(bn128.unserialize(party).mul(r))
		//);
		var temp *ebigint.NBigInt
		if i == p.Index[0] {
			temp = ebigint.NewNBigInt(-int64(p.Value)).ForceRed(b128.Q())
		} else {
			if i == p.Index[1] {
				temp = ebigint.NewNBigInt(int64(p.Value)).ForceRed(b128.Q())
			} else {
				temp = ebigint.NewNBigInt(0).ForceRed(b128.Q())
			}
		}
		log.Println("i == ", i, "temp = ", temp.Text(16))
		t1 := b128.UnSerialize(party).Mul(r)
		log.Println("t1 --->", i, " = ", b128.Serialize(t1))
		log.Println("g = ", b128.Serialize(b128.CurveG()))
		log.Println("g*temp = ", b128.Serialize(b128.CurveG().Mul(temp)))
		C[i] = b128.CurveG().Mul(temp).Add(t1)
		log.Println("C ---> ", i, " = ", b128.Serialize(C[i]))
	}
	var D = b128.CurveG().Mul(r)
	log.Println("D = ", b128.Serialize(D))
	var CLn = make([]types.Point, len(unserialized))
	var CRn = make([]types.Point, len(unserialized))
	for i, account := range unserialized {
		CLn[i] = b128.Serialize(account[0].Add(C[i]))
		CRn[i] = b128.Serialize(account[1].Add(D))
		log.Println("Cln ---> ", i, " = ", CLn[i])
		log.Println("CRn ---> ", i, " = ", CRn[i])
	}

	var NC = make([]types.Point, len(C))
	for i, pc := range C {
		NC[i] = b128.Serialize(pc)
	}

	var ND = b128.Serialize(D)
	var statement core.TransferStatement
	statement.Epoch = p.Epoch
	statement.Y = p.Y
	statement.D = ND
	statement.C = NC
	statement.CLn = CLn
	statement.CRn = CRn

	var witness core.TransferWitness
	witness.Index = p.Index
	witness.BDiff = p.Diff
	witness.BTransfer = p.Value
	witness.R = r.Text(16)
	witness.SK = p.SK
	var proof = core.ProveTransfer(statement, witness)
	log.Println("Proof = ", proof)

	sk := ebigint.FromHex(p.SK)
	var u = b128.Serialize(core.U(p.Epoch, sk))

	type Response struct {
		C     []types.Point `json:"C"`
		D     types.Point   `json:"D"`
		U     types.Point   `json:"u"`
		Y     []types.Point `json:"y"`
		Proof string        `json:"proof"`
	}
	var res Response
	res.C = NC
	res.D = ND
	res.U = u
	res.Y = p.Y
	res.Proof = proof

	b, _ := json.Marshal(res)
	return string(b)
}

func Some(accounts [][2]core.Point) bool {
	var count = 0
	for _, account := range accounts {
		if account[0].Equal(b128.Zero()) && account[1].Equal(b128.Zero()) {
			count += 1
		}
		if count > 1 {
			return true
		}
	}
	return false
}

type BurnProofParam struct {
	Accounts []types.Point `json:"accounts"`
	Epoch    int           `json:"epoch"`
	Value    int           `json:"value"`
	Diff     int           `json:"diff"`
	SK       string        `json:"sk"`
	Y        types.Point   `json:"y"`
	Sender   string        `json:"sender"`
}

func BurnProof(param string) string {
	var p BurnProofParam
	if e := json.Unmarshal([]byte(param), &p); e != nil {
		log.Printf("unmarshal to BurnProofParam failed, err:%s\n", e.Error())
		return ""
	}
	var simulated = p.Accounts
	var CLn = b128.Serialize(b128.UnSerialize(simulated[0]).Add(b128.CurveG().Mul(ebigint.NewNBigInt(-int64(p.Value)))))
	var CRn = simulated[1]
	var statement core.BurnStatement
	statement.Y = p.Y
	statement.Epoch = p.Epoch
	statement.CRn = CRn
	statement.CLn = CLn
	statement.Sender = p.Sender

	var witness core.BurnWitness
	witness.SK = p.SK
	witness.BDiff = p.Diff
	var proof = core.ProveBurn(statement, witness)
	sk := ebigint.FromBytes(common.FromHex(p.SK))
	var u = b128.Serialize(core.U(p.Epoch, sk))

	type Response struct {
		U     types.Point `json:"u"`
		Proof string      `json:"proof"`
	}
	var res Response
	res.U = u
	res.Proof = proof

	b, _ := json.Marshal(res)
	return string(b)
}
