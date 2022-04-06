package main

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

const (
	fromPrivKeyHexstring          = "0xb1b862df61c87139ed6d491b99a0a275fe69fd68b9765a4a442badb2cf2e8358" // 5GppzBbkybS2xtWCeq1W3uLhqEQUk5ZYDjtfMjTAUDbvYboo Csknk
	fromAddress                   = "5GppzBbkybS2xtWCeq1W3uLhqEQUk5ZYDjtfMjTAUDbvYboo"                   // Csknk
	localRecipient                = "0x8eaf04151687736326c9fea17e25fc5287613693c912909cb226aa4794f26a48" // 14E5nqKAp3oAJcmzgZhUD2RcptBeUBScxKHgJKU4HPNcKVf3
	westendRecipientPubkey        = "0x725b16b586c386cf524b067a0449eeef5efc20585f46fe1783db79f1c7cca101" // 5EeeNhoYmB8QKRJ1ffimtb5trLP3bG7gyc6B1cNcnBQCPXH2
	csknkTest2                    = "0xc526d8efca9e85fdce82c6ee694b9690e16c5552b9d276dbe7fe43f7607d4c09" // 5GXCq8BcEzNrmqQN5avx3wARnMBcRyMJzJ5BeT9WrDukpErh"
	inclusionFee           uint64 = 0.0156 * 1e12
)

func ex1() {
	toPubKeyHexstring := csknkTest2 //westendRecipient
	fromPrivKey := fromPrivKeyHexstring
	cfg := "https://westend-rpc.polkadot.io"

	nc, err := NewConnection(cfg)
	if err != nil {
		log.Fatal(err)
	}
	api := nc.Api
	senderPubkey, err := PublicKeyFromAddress(fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	availableBalance, err := nc.GetBalance(hex.EncodeToString(senderPubkey))
	if err != nil {
		log.Fatal(err)
	}
	//	var availableBalance float64 = 0.1607
	var maxSpendable uint64 = availableBalance.Uint64() - inclusionFee
	//	amount := types.NewUCompactFromUInt(westendToBase(maxSpendable))
	amount := types.NewUCompactFromUInt(maxSpendable)

	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		panic(err)
	}

	networkID := uint8(0)
	fromKey, err := signature.KeyringPairFromSecret(fromPrivKey, networkID)

	if err != nil {
		panic(err)
	}

	// Get the nonce for Alice
	to, err := types.NewMultiAddressFromHexAccountID(toPubKeyHexstring)
	if err != nil {
		panic(err)
	}

	// c, err := types.NewCall(meta, "Balances.transfer", to, amount)
	c, err := NewCall(BalanceTransferCallIndex, to, amount)
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
		BlockHash:          genesisHash,
		Era:                types.ExtrinsicEra{IsMortalEra: false},
		GenesisHash:        genesisHash,
		Nonce:              types.NewUCompactFromUInt(uint64(nonce)),
		SpecVersion:        rv.SpecVersion,
		Tip:                types.NewUCompactFromUInt(0),
		TransactionVersion: rv.TransactionVersion,
	}

	fmt.Printf("Sending %v from %#x to %#x with nonce %d\n", amount, fromKey.PublicKey, to.AsID, nonce)

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
