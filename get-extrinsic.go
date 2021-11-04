package main

import (
	"fmt"

	"github.com/centrifuge/go-substrate-rpc-client/v3/types"
	"github.com/vedhavyas/go-subkey"
)

var (
	blockHashStringWestend = "0xd4cba21f5ee0078c21af2e5887cd3c7248ee1ed74bb938530dd034361859a456"
	blockHashString        = "0xd8030c1a1cdb40f8c53f6a0d1db9e6b63ff500ddeda7c6074100d2012689f387"
)

func (c *Connection) getExtrinsic(blockHashStr string, index uint8) {
	// Make a types.Hash object from a hexstring
	blockHash, err := types.NewHashFromHexString(blockHashString)
	if err != nil {
		panic(err)
	}

	// Get the block
	block, err := c.Api.RPC.Chain.GetBlock(blockHash)
	if err != nil {
		panic(err)
	}
	for i, extrinsic := range block.Block.Extrinsics {
		h, _ := types.GetHash(extrinsic)
		fmt.Printf("%#x\n", h)

		who := extrinsic.Signature.Signer.AsID
		fmt.Printf("extrinsic %d signed by: %#x\n", i, who)
	}
	meta, _ := c.Api.RPC.State.GetMetadataLatest()
	key, _ := types.CreateStorageKey(meta, "System", "Events", nil, nil)

	raw, _ := c.Api.RPC.State.GetStorageRaw(key, blockHash)

	events := types.EventRecords{}
	err = types.EventRecordsRaw(*raw).DecodeEventRecords(meta, &events)
	if err != nil {
		panic(err)
	}
	for _, event := range events.Balances_Transfer {
		send, _ := subkey.SS58Address(event.From[:], 0) // 0 is the network identifier byte
		to, _ := subkey.SS58Address(event.To[:], 0)
		fmt.Printf("from: %+v\n", send)
		fmt.Printf("to: %+v\n", to)
		fmt.Printf("value : %+v\n", event.Value)
		fmt.Printf("phase : %+v\n", event.Phase)
		fmt.Printf("topics : %+v\n", event.Topics)

		ext := block.Block.Extrinsics[int(event.Phase.AsApplyExtrinsic)]

		//		fmt.Printf("ext : %+v\n", ext)

		fmt.Printf("nonce : %+v\n", ext.Signature.Nonce)

		fmt.Printf("tip : %+v\n", ext.Signature.Tip)

		extBytes, err := types.EncodeToHexString(ext)
		if err != nil {
			panic(err)
		}
		fmt.Println(extBytes)

		resInter := Fee{}
		err = c.Api.Client.Call(&resInter, "payment_queryInfo", ext, blockHash.Hex())
		if err != nil {
			panic(err)
		}

		fmt.Println("PartialFee: ", resInter.PartialFee)
	}
}

type Fee struct {
	Weight     types.Weight
	Class      string
	PartialFee string
}
