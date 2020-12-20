package client

import (
	"fmt"
	"testing"
)

func TestCreateAccount(t *testing.T) {
	account := CreateAccount("")
	fmt.Println(account)

	specialAccount := CreateAccount("")
	fmt.Println(specialAccount)
}
