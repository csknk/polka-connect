package main

import (
	"fmt"
	"log"

	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
)

// Used for testing purposes - this public key is deterministically generated when a local dev network is
// started using the "alice" flag like so:
// ./polkadot --chain=polkadot-local --alice -d /tmp/alice --node-key 0000000000000000000000000000000000000000000000000000000000000001
const (
	AlicePubkey      string = "0xd43593c715fdd31c61141abd04a99fd6822c8558854ccde39a5684e7a56da27d" // 15oF4uVJwmo4TdGW7VfQxNLavjCXviqxT9S1MgbjMNHr6Sp5
	BobPubkey        string = "0x8eaf04151687736326c9fea17e25fc5287613693c912909cb226aa4794f26a48" // 14E5nqKAp3oAJcmzgZhUD2RcptBeUBScxKHgJKU4HPNcKVf3
	CsknkPubkey      string = "0x627043fea11cb7f7bbeb43bea5a3eecbfd09d9fb5ac9c684cd0b6d73d0802b79"
	TestnetAddr      string = "0x6b1ef2a757b954ed536b96662cba4a5f82cba209451aafe7810d1b09e44c2462" // 5EVAB4TvfpouHMG7EoHbQNfmecoWb5FFJnNX3ab9mMdZhv9H
	TestAddr         string = "0x3e33e5b0cb049ab36ed75f1ab83baf81a2fc5d5bb6d2f6c3283642a49b155d13"
	WestendRecipient        = "0x725b16b586c386cf524b067a0449eeef5efc20585f46fe1783db79f1c7cca101" // 5EeeNhoYmB8QKRJ1ffimtb5trLP3bG7gyc6B1cNcnBQCPXH2
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	// For local dev testnet use NewDefaultConnection()
	//	nc, err := NewConnection("wss://westend-rpc.polkadot.io")
	//	nc, err := NewConnection("wss://rpc.pinknode.io/westend/explorer")
	nc, err := NewDefaultConnection()
	if err != nil {
		log.Fatal(err)
	}

	// To test from a sender other than Alice (the local dev-net auto generated testing identitiy) export
	// in an ENV variable when running:
	// `TEST_PRIV_KEY=0xe7752181c3ff3350c635b0ed640c076d0474642904b9f9cbfabf83a091a673bc ./polka-connect`
	// or export the TEST_PRIV_KEY variable in the shell session and run the script normally.
	sender, ok := signature.LoadKeyringPairFromEnv()
	if !ok {
		sender = signature.TestKeyringPairAlice
	}

	fmt.Println("sender: ", sender.Address)

	//	if err := nc.Transfer(sender, WestendRecipient, dotToPlank(1)); err != nil {
	//		log.Fatal(err)
	//	}

	//	results, err := nc.ChangedBlockHashes(BobPubkey, 0)
	results, err := nc.GetChangeData(BobPubkey, 0)
	if err != nil {
		log.Fatal(err)
	}
	for _, result := range results {

		fmt.Printf("ID: %s\nBlock: %#x\nAmount: %s\n", result.ID, result.blockHash, result.amountAtThisBlock.String())

		//				fmt.Println("blockHash: ", result.Block.Hex())
		//				for _, change := range result.Changes {
		//					fmt.Printf("key: %v\n", change.StorageKey)
		//					var accountInfo AccountInfo
		//					types.DecodeFromBytes(change.StorageData, &accountInfo)
		//					//			decoder.Decode(&accountInfo)
		//
		//					fmt.Printf("accountInfo: %s\n", accountInfo.Data.Free)
		//
		//				}
	}

	/*
		health, err := nc.HealthReportTimeout(1)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("connected in time...")
			fmt.Println("health: ", *health)

		}
		num, err := nc.GetBalance(AlicePubkey)
		//	num, err := nc.GetBalance(TestnetAddr)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("balance: %s\n", num)
		meta, err := nc.getLatestMetadata()
		if err != nil {
			log.Fatal(err)
		}
		m, err := types.EncodeToBytes(meta)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("size of metadata: %d\n", len(m))
	*/

}
