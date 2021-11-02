package main

import (
	"fmt"

	"github.com/centrifuge/go-substrate-rpc-client/v3/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v3/types"
)

func (c *Connection) Transfer(from, to string, amount uint64) error {

	meta, err := c.Api.RPC.State.GetMetadataLatest()
	if err != nil {
		return fmt.Errorf("fetch metadata failed: %w", err)
	}

	// recipient is a MultiAddress struct which will be used to build a suitable Polkadot MultiAddress type.
	// In out case, this will generally be a MultiAddress struct with fields set for `AsID` - containing
	// the public key bytes and `IsID` - a boolean indicating the type of this MultiAddress.
	recipient, err := types.NewMultiAddressFromHexAccountID(to)
	if err != nil {
		return fmt.Errorf("recipient set: %w", err)
	}

	call, err := types.NewCall(meta, "Balances.transfer", recipient, types.NewUCompactFromUInt(amount))
	if err != nil {
		return fmt.Errorf("problem building new call: %w", err)
	}
	extrinsic := types.NewExtrinsic(call)

	genesisHash, err := c.Api.RPC.Chain.GetBlockHash(0)
	if err != nil {
		return fmt.Errorf("failed to get block hash: %w", err)
	}

	runtimeVersion, err := c.Api.RPC.State.GetRuntimeVersionLatest()
	if err != nil {
		return fmt.Errorf("problem getting latest version of runtime: %w", err)
	}

	// From this KeyPair (Alice)
	fr, ok := signature.LoadKeyringPairFromEnv()
	if !ok {
		fr = signature.TestKeyringPairAlice
	}

	// Build a key that will be used to fetch account balance
	key, err := types.CreateStorageKey(meta, "System", "Account", fr.PublicKey)
	if err != nil {
		return fmt.Errorf("problem creating storage key: %w", err)
	}

	var accountInfo types.AccountInfo

	// Alice's account info
	ok, err = c.Api.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil || !ok {
		return fmt.Errorf("problem getting accountInfo: %w", err)
	}

	// Existing on-chain nonce held against this account
	nonce := uint32(accountInfo.Nonce)

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
		return fmt.Errorf("problem creating extrinsic payload: %w", err)
	}

	signature, err := signPayload(payload, fr)
	if err != nil {
		return fmt.Errorf("error signing payload: %w", err)
	}
	era := o.Era
	if !o.Era.IsMortalEra {
		era = types.ExtrinsicEra{IsImmortalEra: true}
	}

	signerPubKey := types.NewMultiAddressFromAccountID(fr.PublicKey)
	fullSignature := types.ExtrinsicSignatureV4{
		Signer:    signerPubKey,
		Signature: types.MultiSignature{IsSr25519: true, AsSr25519: signature},
		Era:       era,
		Nonce:     o.Nonce,
		Tip:       o.Tip,
	}

	extrinsic.Signature = fullSignature
	// mark the extrinsic as signed
	extrinsic.Version |= types.ExtrinsicBitSigned

	// TODO Check this <---------------------------------------------------------------------------------
	//	if err := extrinsic.Sign(fr, o); err != nil {
	//		return fmt.Errorf("problem signing: %w", err)
	//	}
	// <-------------------------------------------------------------------------------------------------

	if extrinsic.IsSigned() {
		fmt.Println("extrinsic is signed")
	}
	hash, err := c.Api.RPC.Author.SubmitExtrinsic(extrinsic)
	if err != nil {
		return fmt.Errorf("submit extrinsic: %w", err)
	}
	fmt.Printf("Transfer sent with hash %#x\n", hash)

	return nil
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
