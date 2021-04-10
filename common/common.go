package common

import (
	"encoding/hex"
	"strconv"
)

var (
	base = "0000000000000000000000000000000000000000000000000000000000000000"
)

func Has0xPrefix(str string) bool {
	return len(str) >= 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X')
}

func FromHex(s string) []byte {
	if Has0xPrefix(s) {
		s = s[2:]
	}
	if len(s)%2 == 1 {
		s = "0" + s
	}
	h, _ := hex.DecodeString(s)
	return h
}
func HexWithout0x(str string) string {
	if Has0xPrefix(str) {
		return str[2:]
	}
	return str
}

func Uint642Bytes32(i uint64) string {
	enc := make([]byte, 0)
	str := string(strconv.AppendUint(enc, i, 16))
	res := base[:len(base)-len(str)] + str
	return res
}
