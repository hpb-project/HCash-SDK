package types

import (
	"encoding/json"
)

type Publickey [2]string

func (p Publickey) GX() string {
	return p[0]
}

func (p Publickey) GY() string {
	return p[1]
}

func (p Publickey) Set(xy []string) Publickey {
	p[0] = xy[0]
	p[1] = xy[1]
	return p
}

func (p Publickey) MarshalJSON() ([]byte, error) {
	type IPubkey struct {
		GX string `json:"gx"`
		GY string `json:"gy"`
	}
	var enc IPubkey
	enc.GX = p[0]
	enc.GY = p[1]
	return json.Marshal(enc)
}

func (p Publickey) String() string {
	d, _ := json.Marshal(p)
	return string(d)
}

type Account struct {
	X string    `json:"x"`
	Y Publickey `json:"y"`
}
