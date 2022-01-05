package main

import (
	"bytes"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/btcsuite/btcutil/base58"
	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/config"
	"github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"golang.org/x/crypto/blake2b"
)

const SS58Prefix = "SS58PRE"

func PublicKeyFromAddress(address string) (publicKey []byte, err error) {
	addressChecksum, publicKey, networkByte, err := addressComponents(address)
	if err != nil {
		return
	}

	computedChecksum, err := Checksum(publicKey, networkByte)
	if err != nil {
		return
	}

	if !bytes.Equal(computedChecksum[:2], addressChecksum) {
		err = fmt.Errorf("invalid checksum for address %s", address)
		return
	}
	return publicKey, nil
}

func Checksum(publicKey []byte, networkByte uint8) ([]byte, error) {
	base := append([]byte{networkByte}, publicKey...)
	hasher, err := blake2b.New(64, nil)
	if err != nil {
		return nil, err
	}
	_, err = hasher.Write(append([]byte(SS58Prefix), base...))
	if err != nil {
		return nil, err
	}
	return hasher.Sum(nil), nil
}

func addressComponents(address string) (addressChecksum, publicKey []byte, networkByte uint8, err error) {
	addressBytes := base58.Decode(address)
	addressChecksum = addressBytes[len(addressBytes)-2:]
	publicKey = addressBytes[1 : len(addressBytes)-2]
	networkByte = addressBytes[0]
	len := len(publicKey)
	if len != 32 {
		err = fmt.Errorf("SS58 address %s yielded wrong length (%d) public key", address, len)
		return nil, nil, 0, err
	}
	return
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

func dotToPlank(dotInput int) uint64 {
	return uint64(dotInput * (1e10))
}

func (c *Connection) GetBlockTimestamp(block *types.SignedBlock, blockHash types.Hash) (*time.Time, error) {

	meta, err := c.getMetadata(blockHash)
	if err != nil {
		return nil, err
	}

	callIndex, err := meta.FindCallIndex("Timestamp.set")
	if err != nil {
		return nil, err
	}

	timestamp := new(big.Int)
	for _, extrinsic := range block.Block.Extrinsics {
		if extrinsic.Method.CallIndex != callIndex {
			continue
		}
		timeDecoder := scale.NewDecoder(bytes.NewReader(extrinsic.Method.Args))
		timestamp, err = timeDecoder.DecodeUintCompact()
		if err != nil {
			return nil, err
		}
		break
	}

	time := time.UnixMilli(timestamp.Int64())
	return &time, nil
}
