package main


import "github.com/hpb-project/HCash-SDK/core/utils/service"

func main() {
	a := service.CreateAccount("123456")
	println("a=",a.String())
}