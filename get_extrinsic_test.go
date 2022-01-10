package main

import (
	"fmt"
	"testing"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

const (
	blockHashStr    = "0x51b3569c9e21b600c689fbe68d56ceaf8f157e0742c6110dc7664686e0333b6b"
	receiverPubKey  = "0x8eaf04151687736326c9fea17e25fc5287613693c912909cb226aa4794f26a48" // 14E5nqKAp3oAJcmzgZhUD2RcptBeUBScxKHgJKU4HPNcKVf3 Bob
	receiverAddress = "14E5nqKAp3oAJcmzgZhUD2RcptBeUBScxKHgJKU4HPNcKVf3"
)

func TestFilterBlockForRequiredExtrinsics(t *testing.T) {
	c, err := NewDefaultConnection()
	if err != nil {
		fmt.Println("No connection to node")
		assert.FailNow(t, err.Error())
	}

	blockHash, err := types.NewHashFromHexString(blockHashStr)
	assert.NoError(t, err)
	accountID, err := types.HexDecodeString(receiverPubKey)
	assert.NoError(t, err)

	err = c.GetRequiredExtrinsicsFromBlockHash([]byte(blockHash[:]), accountID)
	assert.NoError(t, err)
}

func TestHash(t *testing.T) {
	hash, err := types.NewHashFromHexString(blockHashStr)
	assert.NoError(t, err)

	rawBytes := hash[:]
	fmt.Printf("%#x\n", rawBytes)

}

func TestGetTxEvents(t *testing.T) {
	c, err := NewDefaultConnection()
	if err != nil {
		fmt.Println("No connection to node")
		assert.FailNow(t, err.Error())
	}

	blockHash, err := types.NewHashFromHexString(blockHashStr)
	assert.NoError(t, err)

	err = c.GetTxEvents([]byte(blockHash[:]), receiverAddress)
	assert.NoError(t, err)

}
