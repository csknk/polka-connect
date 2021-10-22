package main

import (
	"fmt"
	"log"

	"github.com/centrifuge/go-substrate-rpc-client/config"
	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v3"
)

// Used for testing purposes - this public key is deterministically generated when a local dev network is
// started using the "alice" flag like so:
// ./polkadot --chain=polkadot-local --alice -d /tmp/alice --node-key 0000000000000000000000000000000000000000000000000000000000000001
const ALICE_PUBKEY string = "0xd43593c715fdd31c61141abd04a99fd6822c8558854ccde39a5684e7a56da27d"

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	nc, err := NewConnection()
	if err != nil {
		log.Fatal(err)
	}
	//	nc.HealthReport()
	num, err := nc.GetBalance(ALICE_PUBKEY)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("balance: %s\n", num)

	from := ""
	to := "0x8eaf04151687736326c9fea17e25fc5287613693c912909cb226aa4794f26a48"
	if err := nc.Transfer(from, to, 4200); err != nil {
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
