package main

import (
	"fmt"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/config"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

func ex1() {
	fromPrivKeyHexstring := "0xb1b862df61c87139ed6d491b99a0a275fe69fd68b9765a4a442badb2cf2e8358" // Csknk
	// fromAddr := "15m88WrpqNhWQRWicU4WC4ArgrQ8SP7gJEd9X2SX2JdSiyRc"
	//	toAddress := "14E5nqKAp3oAJcmzgZhUD2RcptBeUBScxKHgJKU4HPNcKVf3" // Bob
	toPubKeyHexstring := "" // Bob

	// Display the events that occur during a transfer by sending a value to bob

	// Instantiate the API
	//	api, err := gsrpc.NewSubstrateAPI("https://westend-rpc.polkadot.io")
	cfg := config.Default().RPCURL
	api, err := gsrpc.NewSubstrateAPI(cfg)
	if err != nil {
		panic(err)
	}

	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		panic(err)
	}

	// Create a call, transferring 12345 units to Bob
	//	from, err := types.NewMultiAddressFromHexAccountID("0x1f4b81480f9fc66e2e1e6db4849bf7dc0b5fbfe68e165d5e13178fe8af0a9d15")
	//	from, err := types.NewMultiAddressFromHexAccountID(fromPubKeyHexstring)
	//	if err != nil {
	//		panic(err)
	//	}

	//	fromPrivKey := "0x1f4b81480f9fc66e2e1e6db4849bf7dc0b5fbfe68e165d5e13178fe8af0a9d15"
	fromPrivKey := fromPrivKeyHexstring
	netwrokId := uint8(0)
	fromKey, err := signature.KeyringPairFromSecret(fromPrivKey, netwrokId)

	if err != nil {
		panic(err)
	}

	amount := types.NewUCompactFromUInt(dotToPlank(2))

	// Get the nonce for Alice
	//	to, err := types.NewMultiAddressFromHexAccountID("0x4c4f0e86470be8bce081440c8b9cb2703bee894340173775442ae123d4fe1b71")
	to, err := types.NewMultiAddressFromHexAccountID(toPubKeyHexstring)
	if err != nil {
		panic(err)
	}

	c, err := types.NewCall(meta, "Balances.transfer", to, amount)
	if err != nil {
		panic(err)
	}

	// Create the extrinsic
	ext := types.NewExtrinsic(c)
	if err != nil {
		panic(err)
	}

	genesisHash, err := api.RPC.Chain.GetBlockHash(0)
	if err != nil {
		panic(err)
	}

	rv, err := api.RPC.State.GetRuntimeVersionLatest()
	if err != nil {
		panic(err)
	}

	// Build a key that will be used to fetch account balance for the sending account
	//	key, err := types.CreateStorageKey(meta, "System", "Account", from.AsID[:], nil)
	key, err := types.CreateStorageKey(meta, "System", "Account", fromKey.PublicKey, nil)
	if err != nil {
		fmt.Printf("failed to create storage key: %v", err)
		return
	}

	// NOTE: decoding data into gsrpc/types.AccountInfo will provide wrong result.
	var accountInfo types.AccountInfo
	//	_, err = api.RPC.State.GetStorageLatest(key, &accountInfo)
	//	if err != nil {
	//		fmt.Printf("problem getting accountInfo: %v", err)
	//		return
	//	}
	// Sender's account info - put data into accountInfo
	ok, err := api.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil || !ok {
		fmt.Printf("problem getting accountInfo: %v", err)
		return
	}

	// Underlying data type for the nonce is uint32 - keep it as this and let callers cast it if required.
	nonce := uint32(accountInfo.Nonce)

	o := types.SignatureOptions{
		BlockHash:   genesisHash,
		Era:         types.ExtrinsicEra{IsMortalEra: false},
		GenesisHash: genesisHash,
		Nonce:       types.NewUCompactFromUInt(uint64(nonce)),
		SpecVersion: rv.SpecVersion,
		Tip:         types.NewUCompactFromUInt(0),
		// Necessary:
		TransactionVersion: rv.TransactionVersion,
	}

	fmt.Printf("Sending %v from %#x to %#x with nonce %v", amount, fromKey.PublicKey, to.AsID, nonce)

	// Sign the transaction using Alice's default account
	//	err = ext.Sign(fromKey, o)
	//	if err != nil {
	//		panic(err)
	//	}

	// --------------------------------------------------------------------------------------------------
	// Unsigned Payload
	payload, err := createUnsignedPayload(&ext, o)
	if err != nil {
		fmt.Printf("problem creating extrinsic payload: %v", err)
		return
	}

	signature, err := signPayload(payload, fromKey)
	if err != nil {
		fmt.Printf("error signing payload: %v", err)
		return
	}

	era := o.Era
	if !o.Era.IsMortalEra {
		era = types.ExtrinsicEra{IsImmortalEra: true}
	}

	// Signer must be in MultiAddress format
	signerPubKey := types.NewMultiAddressFromAccountID(fromKey.PublicKey)
	fullSignature := types.ExtrinsicSignatureV4{
		Signer:    signerPubKey,
		Signature: types.MultiSignature{IsSr25519: true, AsSr25519: signature},
		Era:       era,
		Nonce:     o.Nonce,
		Tip:       o.Tip,
	}

	ext.Signature = fullSignature

	// mark the extrinsic as signed - extrinsic.IsSigned will now return true
	ext.Version |= types.ExtrinsicBitSigned

	/*
		payloadBytes, err := types.EncodeToBytes(payload)
		if err != nil {
			return nil, nil, err
		}
	*/

	tx, err := api.RPC.Author.SubmitExtrinsic(ext)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", tx.Hex())
	// // Do the transfer and track the actual status
	// sub, err := api.RPC.Author.SubmitAndWatchExtrinsic(ext)
	// if err != nil {
	// 	panic(err)
	// }
	// defer sub.Unsubscribe()

	// for {
	// 	status := <-sub.Chan()
	// 	fmt.Printf("Transaction status: %#v\n", status)

	// 	if status.IsInBlock {
	// 		fmt.Printf("Completed at block hash: %#x\n", status.AsInBlock)
	// 		return
	// 	}
	// }
}

func dotToPlank(dotInput int) uint64 {
	return uint64(dotInput * (1e10))
}