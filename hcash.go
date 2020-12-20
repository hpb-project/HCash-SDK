package main

import (
	"github.com/hpb-project/HCash-SDK/core/utils"
)

func main() {
	a := utils.CreateAccount("123456")
	println("a=", a.String())
}
