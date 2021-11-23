package main

import (
	"bytes"
	"fmt"
	"math/big"
	"testing"

	"github.com/centrifuge/go-substrate-rpc-client/v3/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v3/types"
	"github.com/stretchr/testify/assert"
)

func TestEncodeDecodeExtrinsic(t *testing.T) {
	c, err := NewDefaultConnection()
	if err != nil {
		fmt.Println("No connection to node")
		assert.FailNow(t, err.Error())
	}
	AlicePubkey := "0xd43593c715fdd31c61141abd04a99fd6822c8558854ccde39a5684e7a56da27d" // 15oF4uVJwmo4TdGW7VfQxNLavjCXviqxT9S1MgbjMNHr6Sp5
	BobPubkey := "0x8eaf04151687736326c9fea17e25fc5287613693c912909cb226aa4794f26a48"   // 14E5nqKAp3oAJcmzgZhUD2RcptBeUBScxKHgJKU4HPNcKVf3
	var amount uint64 = 4200000000

	extrinsic, err := c.NewExtrinsic(AlicePubkey, BobPubkey, amount)
	assert.NoError(t, err)

	extrinsicString, err := types.EncodeToHexString(extrinsic)
	if err != nil {
		assert.NoError(t, err)
	}
	fmt.Println("Signed extrinsic: ", extrinsicString)

	_ = DecodeExtrinsic(t, c, extrinsicString)

}

func DecodeExtrinsic(t *testing.T, c *Connection, extrinsicString string) error {
	var decodedExtrinsic types.Extrinsic
	if err := types.DecodeFromHexString(extrinsicString, &decodedExtrinsic); err != nil {
		assert.Fail(t, err.Error())
	}
	fmt.Printf("%+v\n", decodedExtrinsic.Method.Args)

	decoder := scale.NewDecoder(bytes.NewReader(decodedExtrinsic.Method.Args))

	// Determine number of calls
	// NOTE IT IS NECESSARY TO REMOVE BYTE FROM STREAM - otherwise subsequent decoding won't work properly.
	n, err := decoder.DecodeUintCompact()
	if err != nil {
		assert.Fail(t, err.Error())
	}
	fmt.Println("CALLS # ", n)

	// TODO	Maybe loop over n - can there be multiple calls?
	metadata, _ := c.getLatestMetadata()
	//	fmt.Printf("%#x\n", decodedExtrinsic.Signature.Signer.AsID)

	callFunction := findModule(metadata, decodedExtrinsic.Method.CallIndex)

	for _, callArg := range callFunction.Args {
		if callArg.Type == "<T::Lookup as StaticLookup>::Source" {
			var argValue = types.AccountID{}
			decoder.Decode(&argValue)
			fmt.Println(callArg.Name, " = ", argValue)
			fmt.Printf("%#x\n", argValue)
		}
		if callArg.Type == "Compact<T::Balance>" {
			amount, _ := decoder.DecodeUintCompact()
			fmt.Println(callArg.Name, " = ", amount)
		}
		if callArg.Type == "Vec<u8>" {
			var argValue = callArg.Name
			_ = decoder.Decode(&argValue)
			fmt.Println(callArg.Name, " = ", argValue)
		}
	}
	return nil
}

func findModule(metadata *types.Metadata, index types.CallIndex) types.FunctionMetadataV4 {
	for _, mod := range metadata.AsMetadataV13.Modules {
		if mod.Index == index.SectionIndex {
			return mod.Calls[index.MethodIndex]
		}
	}
	panic("Unknown call")
}

var ExamplaryExtrinsic = types.Extrinsic{Version: 0x84, Signature: types.ExtrinsicSignatureV4{Signer: types.MultiAddress{IsID: true, AsID: types.AccountID{0xd4, 0x35, 0x93, 0xc7, 0x15, 0xfd, 0xd3, 0x1c, 0x61, 0x14, 0x1a, 0xbd, 0x4, 0xa9, 0x9f, 0xd6, 0x82, 0x2c, 0x85, 0x58, 0x85, 0x4c, 0xcd, 0xe3, 0x9a, 0x56, 0x84, 0xe7, 0xa5, 0x6d, 0xa2, 0x7d}}, Signature: types.MultiSignature{IsSr25519: true, AsSr25519: types.Signature{0x5c, 0x77, 0x1d, 0xd5, 0x6a, 0xe0, 0xce, 0xed, 0x68, 0xd, 0xb3, 0xbb, 0x4c, 0x40, 0x7a, 0x38, 0x96, 0x99, 0x97, 0xae, 0xb6, 0xa, 0x2c, 0x62, 0x39, 0x1, 0x6, 0x2f, 0x7f, 0x8e, 0xbf, 0x2f, 0xe7, 0x73, 0x3a, 0x61, 0x3c, 0xf1, 0x6b, 0x78, 0xf6, 0x10, 0xc6, 0x52, 0x32, 0xa2, 0x3c, 0xc5, 0xce, 0x25, 0xda, 0x29, 0xa3, 0xd5, 0x84, 0x85, 0xd8, 0x7b, 0xd8, 0x3d, 0xb8, 0x18, 0x3f, 0x8}}, Era: types.ExtrinsicEra{IsImmortalEra: true, IsMortalEra: false, AsMortalEra: types.MortalEra{First: 0x0, Second: 0x0}}, Nonce: types.UCompact(*big.NewInt(1)), Tip: types.UCompact(*big.NewInt(2))}, Method: types.Call{CallIndex: types.CallIndex{SectionIndex: 0x3, MethodIndex: 0x0}, Args: types.Args{0xff, 0x8e, 0xaf, 0x4, 0x15, 0x16, 0x87, 0x73, 0x63, 0x26, 0xc9, 0xfe, 0xa1, 0x7e, 0x25, 0xfc, 0x52, 0x87, 0x61, 0x36, 0x93, 0xc9, 0x12, 0x90, 0x9c, 0xb2, 0x26, 0xaa, 0x47, 0x94, 0xf2, 0x6a, 0x48, 0xe5, 0x6c}}}
