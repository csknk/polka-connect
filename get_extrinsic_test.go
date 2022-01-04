package main

import (
	"fmt"
	"testing"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

const (
	blockHashStr   = "0xc69fe03e69a1505131c6a1c61c78848535d3e9584d7159c1e0e37fc864cdea02"
	receiverPubKey = "0x8eaf04151687736326c9fea17e25fc5287613693c912909cb226aa4794f26a48" // 14E5nqKAp3oAJcmzgZhUD2RcptBeUBScxKHgJKU4HPNcKVf3 Bob
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

	err = c.FilterBlockForRequiredExtrinsics([]byte(blockHash[:]), accountID)
	assert.NoError(t, err)
}
