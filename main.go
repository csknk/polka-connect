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
const (
	AlicePubkey string = "0xd43593c715fdd31c61141abd04a99fd6822c8558854ccde39a5684e7a56da27d" // 15oF4uVJwmo4TdGW7VfQxNLavjCXviqxT9S1MgbjMNHr6Sp5
	BobPubkey   string = "0x8eaf04151687736326c9fea17e25fc5287613693c912909cb226aa4794f26a48" // 14E5nqKAp3oAJcmzgZhUD2RcptBeUBScxKHgJKU4HPNcKVf3
	CsknkPubkey string = "0x627043fea11cb7f7bbeb43bea5a3eecbfd09d9fb5ac9c684cd0b6d73d0802b79"
	TestAddr    string = "0x3e33e5b0cb049ab36ed75f1ab83baf81a2fc5d5bb6d2f6c3283642a49b155d13"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	nc, err := NewDefaultConnection()
	if err != nil {
		log.Fatal(err)
	}
	//	nc.HealthReport()

	add, err := nc.GetAddress([]byte{
		0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01,
		0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01,
		0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01,
		0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(add.AsAccountID)

	var buf bytes.Buffer
	e := scale.NewEncoder(&buf)
	add.Encode(*e)
	fmt.Printf("e: %#x\n", &buf)

	//	num, err := nc.GetBalance(AlicePubkey)
	num, err := nc.GetBalance(TestAddr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("balance: %s\n", num)

	//	from := ""
	to := "0x8eaf04151687736326c9fea17e25fc5287613693c912909cb226aa4794f26a48"
	if err := nc.Transfer(AlicePubkey, to, 4200); err != nil {
		log.Fatal(err)
	}
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
