package types

import (
	"encoding/json"
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
