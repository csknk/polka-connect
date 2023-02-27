package main

import (
	"fmt"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/vedhavyas/go-subkey"
)

func ReadExtrinsic(c *Connection, blockHash string) {
	meta, err := c.Api.RPC.State.GetMetadataLatest()
	checkErr(err)

	hash, err := types.NewHashFromHexString(blockHash)
	checkErr(err)

	key, err := types.CreateStorageKey(meta, "System", "Events", nil, nil)
	checkErr(err)

	raw, err := c.Api.RPC.State.GetStorageRaw(key, hash)
	checkErr(err)

	events := types.EventRecords{}
	err = types.EventRecordsRaw(*raw).DecodeEventRecords(meta, &events)
	checkErr(err)

	fmt.Println("Read Block blockHash: ", hash.Hex())

	// Get the block
	block, err := c.Api.RPC.Chain.GetBlock(hash)
	checkErr(err)

	for _, event := range events.Balances_Transfer {
		send, _ := subkey.SS58Address(event.From[:], 0)
		fmt.Printf("from : %+v\n", send)
		to, _ := subkey.SS58Address(event.To[:], 0)
		fmt.Printf("to : %+v\n", to)
		fmt.Printf("value : %+v\n", event.Value)
		fmt.Printf("phase : %+v\n", event.Phase)
		fmt.Printf("topics : %+v\n", event.Topics)

		ext := block.Block.Extrinsics[int(event.Phase.AsApplyExtrinsic)]

		fmt.Printf("ext : %+v\n", ext)

		fmt.Printf("nonce : %+v\n", ext.Signature.Nonce)

		fmt.Printf("tip : %+v\n", ext.Signature.Tip)

		extBytes, err := types.EncodeToHexString(ext)
		checkErr(err)
		fmt.Println(extBytes)

		resInter := Fee{}
		err = c.Api.Client.Call(&resInter, "payment_queryInfo", ext, hash.Hex())
		checkErr(err)

		fmt.Println("PartialFee: ", resInter.PartialFee)
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
