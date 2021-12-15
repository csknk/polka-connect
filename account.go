package main

import (
	"fmt"
	"math/big"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/config"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

func account() {
	// This example shows how to instantiate a Substrate API and use it to connect to a node and retrieve balance
	// updates
	//
	// NOTE: The example runs until you stop it with CTRL+C

	api, err := gsrpc.NewSubstrateAPI(config.Default().RPCURL)
	if err != nil {
		panic(err)
	}

	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		fmt.Println("GetMetaDataLatest error")

		panic(err)
	}

	// Known account we want to use (available on dev chain, with funds)
	alice, err := types.HexDecodeString("0xd43593c715fdd31c61141abd04a99fd6822c8558854ccde39a5684e7a56da27d")
	if err != nil {
		fmt.Println("HexDecodeString error")

		panic(err)
	}

	key, err := types.CreateStorageKey(meta, "Balances", "FreeBalance", alice, nil)
	if err != nil {
		fmt.Println("CreateStorageKey error")

		panic(err)
	}

	// Retrieve the initial balance
	var previous types.U128
	ok, err := api.RPC.State.GetStorageLatest(key, &previous)
	if err != nil || !ok {
		panic(err)
	}

	fmt.Printf("%#x has a balance of %v\n", alice, previous)
	fmt.Printf("You may leave this example running and transfer any value to %#x\n", alice)

	// Here we subscribe to any balance changes
	sub, err := api.RPC.State.SubscribeStorageRaw([]types.StorageKey{key})
	if err != nil {
		panic(err)
	}
	defer sub.Unsubscribe()

	// outer for loop for subscription notifications
	for {
		// inner loop for the changes within one of those notifications
		for _, chng := range (<-sub.Chan()).Changes {
			var current types.U128
			if err = types.DecodeFromBytes(chng.StorageData, &current); err != nil {
				panic(err)
			}

			// Calculate the delta
			var change = types.U128{Int: big.NewInt(0).Sub(current.Int, previous.Int)}

			// Only display positive value changes (Since we are pulling `previous` above already,
			// the initial balance change will also be zero)
			if change.Cmp(big.NewInt(0)) != 0 {
				previous = current
				fmt.Printf("New balance change of: %v\n", change)
				return
			}
		}
	}
}
