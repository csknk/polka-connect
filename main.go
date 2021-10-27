package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/centrifuge/go-substrate-rpc-client/config"
	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v3"
	"github.com/centrifuge/go-substrate-rpc-client/v3/scale"
)

// Used for testing purposes - this public key is deterministically generated when a local dev network is
// started using the "alice" flag like so:
// ./polkadot --chain=polkadot-local --alice -d /tmp/alice --node-key 0000000000000000000000000000000000000000000000000000000000000001
const ALICE_PUBKEY string = "0xd43593c715fdd31c61141abd04a99fd6822c8558854ccde39a5684e7a56da27d"
const CSKNK_PUBKEY string = "0x627043fea11cb7f7bbeb43bea5a3eecbfd09d9fb5ac9c684cd0b6d73d0802b79"

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	nc, err := NewDefaultConnection()
	if err != nil {
		log.Fatal(err)
	}
	//	nc.HealthReport()

	add, err := nc.GetAddress([]byte{0x01, 32})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(add.AsAccountID)

	// create writer
	var buf bytes.Buffer
	dec := scale.NewDecoder(&buf)
	add.Decode(*dec)
	//	fmt.Println("dec: ++++++++++++++")

	fmt.Println(dec)

	//	num, err := nc.GetBalance(ALICE_PUBKEY)
	num, err := nc.GetBalance(CSKNK_PUBKEY)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("balance: %s\n", num)

	//	from := ""
	//	to := "0x8eaf04151687736326c9fea17e25fc5287613693c912909cb226aa4794f26a48"
	//	if err := nc.Transfer(from, to, 4200); err != nil {
	//		log.Fatal(err)
	//	}
	//	printLatestBlockHash()
}

func printLatestBlockHash() {
	api, err := gsrpc.NewSubstrateAPI(config.Default().RPCURL)
	if err != nil {
		log.Fatal(err)
	}
	hash, err := api.RPC.Chain.GetBlockHashLatest()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(hash.Hex())
}
