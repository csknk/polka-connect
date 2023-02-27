package core

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestGetBalance(t *testing.T) {
	nc, err := NewConnection("wss://rpc.polkadot.io")
	if err != nil {
		t.Fatal(err)
	}
	//	id := "0x0231438b10ffedcb77d5d7b16b9a8874a4fbc6d60c9e5b128351f99952ed5190"
	//	id := "0x7fb6a003140b2types2e2f536c6f6447547d283938aae6344434ccb55a1b3ef03bbd4"
	address := "13NRFigKWtUSabcWM8WQ7KJnh1Sqz63sN6iiSFc6TM8h1wcM"
	pubKey, err := PublicKeyFromAddress(address)
	if err != nil {
		t.Fatal(err)
	}

	id := hex.EncodeToString(pubKey)
	balance, nonce, err := nc.GetBalance(id)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%s balance: %d nonce: %d\n", address, balance, nonce)
}
