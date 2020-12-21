package main

import (
	"github.com/hpb-project/HCash-SDK/core"
)

func main() {
	a := core.CreateAccount("123456")
	println("a=", a.String())
}
