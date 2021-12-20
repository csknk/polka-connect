package main

import (
	"fmt"

	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

type Transaction *types.Extrinsic

func (c *Connection) NewExtrinsic(sender signature.KeyringPair, to string, amount uint64) (*types.Extrinsic, error) {

	meta, err := c.Api.RPC.State.GetMetadataLatest()
	if err != nil {
		return nil, fmt.Errorf("fetch metadata failed: %w", err)
	}

	// recipient is a MultiAddress struct which will be used to build a suitable Polkadot MultiAddress type.
	// In our case, this will generally be a MultiAddress struct with fields set for `AsID` - containing
	// the public key bytes and `IsID` - a boolean indicating the type of this MultiAddress.
	recipient, err := types.NewMultiAddressFromHexAccountID(to)
	if err != nil {
		return nil, fmt.Errorf("recipient set: %w", err)
	}

	call, err := types.NewCall(meta, "Balances.transfer", recipient, types.NewUCompactFromUInt(amount))
	if err != nil {
		return nil, fmt.Errorf("problem building new call: %w", err)
	}

	extrinsic := types.NewExtrinsic(call)

	genesisHash, err := c.Api.RPC.Chain.GetBlockHash(0)
	if err != nil {
		return nil, fmt.Errorf("failed to get block hash: %w", err)
	}

	runtimeVersion, err := c.Api.RPC.State.GetRuntimeVersionLatest()
	if err != nil {
		return nil, fmt.Errorf("problem getting latest version of runtime: %w", err)
	}

	fmt.Printf("Sending from:\nPublic key: %#x\nAddress: %s", sender.PublicKey, sender.Address)

	// Build a key that will be used to fetch account balance
	key, err := types.CreateStorageKey(meta, "System", "Account", sender.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("problem creating storage key: %w", err)
	}

	var senderAccountInfo types.AccountInfo

	// Sender's account info
	ok, err := c.Api.RPC.State.GetStorageLatest(key, &senderAccountInfo)
	if err != nil || !ok {
		return nil, fmt.Errorf("problem getting senderAccountInfo: %w", err)
	}

	fmt.Printf("Sending account data\nBalance: %v\nNonce: %v\n",
		senderAccountInfo.Data.Free,
		senderAccountInfo.Nonce)

	// Existing on-chain nonce held against the sending account
	nonce := uint32(senderAccountInfo.Nonce)

	// Set signature options.
	o := types.SignatureOptions{
		BlockHash:          genesisHash,
		Era:                types.ExtrinsicEra{IsMortalEra: false},
		GenesisHash:        genesisHash,
		Nonce:              types.NewUCompactFromUInt(uint64(nonce)),
		SpecVersion:        runtimeVersion.SpecVersion,
		Tip:                types.NewUCompactFromUInt(0),
		TransactionVersion: runtimeVersion.TransactionVersion,
	}

	// Unsigned Payload - note that the entire Extrinsic is not signed, just the payload. The signature
	// is then attached to the Extrinsic, embedded within a types.ExtrinsicSignatureV4.MultiSignature object.
	payload, err := createUnsignedPayload(&extrinsic, o)
	if err != nil {
		return nil, fmt.Errorf("problem creating extrinsic payload: %w", err)
	}

	// signPayload is a placeholder for MPC signing functionality.
	// NOTE GSRPC library can only sign messages using Sr25519 scheme.
	signature, err := signPayload(payload, sender)
	if err != nil {
		return nil, fmt.Errorf("error signing payload: %w", err)
	}

	// watcher, async
	era := o.Era
	if !o.Era.IsMortalEra {
		era = types.ExtrinsicEra{IsImmortalEra: true}
	}

	// Signer must be in MultiAddress format
	signerPubKey := types.NewMultiAddressFromAccountID(sender.PublicKey)
	fullSignature := types.ExtrinsicSignatureV4{
		Signer:    signerPubKey,
		Signature: types.MultiSignature{IsSr25519: true, AsSr25519: signature},
		Era:       era,
		Nonce:     o.Nonce,
		Tip:       o.Tip,
	}

	extrinsic.Signature = fullSignature

	// mark the extrinsic as signed - extrinsic.IsSigned will now return true
	extrinsic.Version |= types.ExtrinsicBitSigned

	return &extrinsic, nil
}

func signPayload(payload types.ExtrinsicPayloadV4, signer signature.KeyringPair) (types.Signature, error) {
	// This is what must be sent to MPC network for signing.
	bytes, err := types.EncodeToBytes(payload)
	if err != nil {
		return types.Signature{}, err
	}

	// Sign data with the private key under the given derivation path - for GSRPC, the signature scheme
	// is set to sr25519. Depending on how the MPC signing works, we will probably just need to add the
	// signature over the message bytes into a Signature struct (see below)
	//
	// NOTE: If data is longer than 256 bytes, hash first:
	// if len(data) > 256 {
	//	h := blake2b.Sum256(data)
	//	data = h[:]
	// }

	sig, err := signature.Sign(bytes, signer.URI)

	// NewSignature just copies signature []byte b into a Signature{} h: copy(h[:], b)
	return types.NewSignature(sig), err
}

// unsignedExtrinsic
func createUnsignedPayload(extrinsic *(types.Extrinsic), options types.SignatureOptions) (types.ExtrinsicPayloadV4, error) {
	payload := types.ExtrinsicPayloadV4{}
	if (*extrinsic).Type() != types.ExtrinsicVersion4 {
		return payload, fmt.Errorf(
			"unsupported extrinsic version: %v (isSigned: %v, type: %v)",
			(*extrinsic).Version,
			(*extrinsic).IsSigned(),
			(*extrinsic).Type())
	}

	mb, err := types.EncodeToBytes((*extrinsic).Method)
	if err != nil {
		return payload, err
	}

	era := options.Era
	if !options.Era.IsMortalEra {
		era = types.ExtrinsicEra{IsImmortalEra: true}
	}

	payload = types.ExtrinsicPayloadV4{
		ExtrinsicPayloadV3: types.ExtrinsicPayloadV3{
			Method:      mb,
			Era:         era,
			Nonce:       options.Nonce,
			Tip:         options.Tip,
			SpecVersion: options.SpecVersion,
			GenesisHash: options.GenesisHash,
			BlockHash:   options.BlockHash,
		},
		TransactionVersion: options.TransactionVersion,
	}

	return payload, nil
}

// unsignedExtrinsic
// This is a function for the time being. Consider embedding parent Extrinsic in a custom Extrinsic struct
// and providing a method to generate an unsigned transaction.
func signExtrinsic(extrinsic *(types.Extrinsic), signer signature.KeyringPair, options types.SignatureOptions) error {
	if (*extrinsic).Type() != types.ExtrinsicVersion4 {
		return fmt.Errorf("unsupported extrinsic version: %v (isSigned: %v, type: %v)", (*extrinsic).Version, (*extrinsic).IsSigned(), (*extrinsic).Type())
	}

	mb, err := types.EncodeToBytes((*extrinsic).Method)
	if err != nil {
		return err
	}

	era := options.Era
	if !options.Era.IsMortalEra {
		era = types.ExtrinsicEra{IsImmortalEra: true}
	}

	payload := types.ExtrinsicPayloadV4{
		ExtrinsicPayloadV3: types.ExtrinsicPayloadV3{
			Method:      mb,
			Era:         era,
			Nonce:       options.Nonce,
			Tip:         options.Tip,
			SpecVersion: options.SpecVersion,
			GenesisHash: options.GenesisHash,
			BlockHash:   options.BlockHash,
		},
		TransactionVersion: options.TransactionVersion,
	}

	signerPubKey := types.NewMultiAddressFromAccountID(signer.PublicKey)

	sig, err := payload.Sign(signer)
	if err != nil {
		return err
	}

	extSig := types.ExtrinsicSignatureV4{
		Signer:    signerPubKey,
		Signature: types.MultiSignature{IsSr25519: true, AsSr25519: sig},
		Era:       era,
		Nonce:     options.Nonce,
		Tip:       options.Tip,
	}

	(*extrinsic).Signature = extSig

	// mark the extrinsic as signed
	(*extrinsic).Version |= types.ExtrinsicBitSigned

	return nil
}

func (c *Connection) GenTransaction(currency int, from, to string, amount uint64) (tx *Transaction, toBeSigned []byte, err error) {
	meta, err := c.Api.RPC.State.GetMetadataLatest()
	if err != nil {
		return nil, nil, fmt.Errorf("fetch metadata failed: %w", err)
	}

	// recipient is a MultiAddress struct which will be used to build a suitable Polkadot MultiAddress type.
	// In our case, this will generally be a MultiAddress struct with fields set for `AsID` - containing
	// the public key bytes and `IsID` - a boolean indicating the type of this MultiAddress.
	recipient, err := types.NewMultiAddressFromHexAccountID(to)
	if err != nil {
		return nil, nil, fmt.Errorf("recipient set: %w", err)
	}

	call, err := types.NewCall(meta, "Balances.transfer", recipient, types.NewUCompactFromUInt(amount))
	if err != nil {
		return nil, nil, fmt.Errorf("problem building new call: %w", err)
	}

	extrinsic := types.NewExtrinsic(call)

	genesisHash, err := c.Api.RPC.Chain.GetBlockHash(0)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get block hash: %w", err)
	}

	runtimeVersion, err := c.Api.RPC.State.GetRuntimeVersionLatest()
	if err != nil {
		return nil, nil, fmt.Errorf("problem getting latest version of runtime: %w", err)
	}

	// Build a key that will be used to fetch account balance
	fromPubKey, err := PublicKeyFromAddress(from)
	if err != nil {
		return nil, nil, err
	}

	key, err := types.CreateStorageKey(meta, "System", "Account", fromPubKey)
	if err != nil {
		return nil, nil, fmt.Errorf("problem creating storage key: %w", err)
	}

	var senderAccountInfo types.AccountInfo

	// Sender's account info
	ok, err := c.Api.RPC.State.GetStorageLatest(key, &senderAccountInfo)
	if err != nil || !ok {
		return nil, nil, fmt.Errorf("problem getting senderAccountInfo: %w", err)
	}

	// Existing on-chain nonce held against this account
	nonce := uint32(senderAccountInfo.Nonce)

	// Set signature options.
	o := types.SignatureOptions{
		BlockHash:          genesisHash,
		Era:                types.ExtrinsicEra{IsMortalEra: false},
		GenesisHash:        genesisHash,
		Nonce:              types.NewUCompactFromUInt(uint64(nonce)),
		SpecVersion:        runtimeVersion.SpecVersion,
		Tip:                types.NewUCompactFromUInt(0),
		TransactionVersion: runtimeVersion.TransactionVersion,
	}

	// Unsigned Payload
	payload, err := createUnsignedPayload(&extrinsic, o)
	if err != nil {
		return nil, nil, fmt.Errorf("problem creating extrinsic payload: %w", err)
	}
	payloadBytes, err := types.EncodeToBytes(payload)
	if err != nil {
		return nil, nil, err
	}

	return nil, payloadBytes, nil
}
