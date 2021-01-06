package types

import (
	"bytes"
	"encoding/json"
	"github.com/hpb-project/HCash-SDK/common"
)

type Point [2]string

func (p Point) GX() string {
	return p[0]
}

func (p Point) GY() string {
	return p[1]
}

func (p Point) Set(xy []string) Point {
	p[0] = xy[0]
	p[1] = xy[1]
	return p
}

func (p Point) Equal(o Point) bool {
	return p[0] == o[0] && p[1] == o[1]
}

func (p Point) MarshalJSON() ([]byte, error) {
	type IPubkey struct {
		GX string `json:"gx"`
		GY string `json:"gy"`
	}
	var enc IPubkey
	enc.GX = p[0]
	enc.GY = p[1]
	return json.Marshal(enc)
}

func (p Point) String() string {
	d, _ := json.Marshal(p)
	return string(d)
}

func (p Point) Match(o Point) bool {
	px := common.FromHex(p.GX())
	py := common.FromHex(p.GY())

	ox := common.FromHex(o.GX())
	oy := common.FromHex(o.GY())

	if bytes.Compare(px, ox) == 0 && bytes.Compare(py, oy) == 0 {
		return true
	} else {
		return false
	}
}
