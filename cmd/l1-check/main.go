package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"polka-connect/core"
)

func main() {
	inFilePath := "manifest.txt"
	f, err := os.Open(inFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	addresses := []string{}
	for sc.Scan() {
		addresses = append(addresses, sc.Text())
	}
	endpoint := "wss://westend-rpc.polkadot.io" // "wss://rpc.polkadot.io"
	nc, err := core.NewConnection(endpoint)
	if err != nil {
		log.Fatal(err)
	}

	layerOneData := []*LayerOneData{}
	for i, address := range addresses {
		data, err := run(nc, address)
		if err != nil {
			log.Fatalf("address %d %s; error running check: %v", i, address, err)
		}
		layerOneData = append(layerOneData, data)
	}
	for _, el := range layerOneData {
		fmt.Printf("%s\n", el)
	}
}

type LayerOneData struct {
	Address string `json:"address"`
	Balance int64  `json:"balance"`
	Nonce   uint32 `json:"nonce"`
}

func (l *LayerOneData) String() string {
	formatString := "%s \tnonce: %d\tbalance(Planck, DOT): %d\t\t%f"
	corrected := float64(l.Balance) / 1e10
	return fmt.Sprintf(formatString, l.Address, l.Nonce, l.Balance, corrected)
}

func run(nc *core.Connection, address string) (*LayerOneData, error) {
	pubKey, err := core.PublicKeyFromAddress(address)
	if err != nil {
		return nil, err
	}

	id := hex.EncodeToString(pubKey)
	balance, nonce, err := nc.GetBalance(id)
	if err != nil {
		return nil, err
	}
	return &LayerOneData{
		Address: address,
		Balance: balance.Int64(),
		Nonce:   uint32(nonce),
	}, nil
}
