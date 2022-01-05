package main

import (
	"fmt"
	"testing"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

const (
	blockHashStr   = "0x581f35cfb19cd46315fce84292b8ce0ff8455c67da8c998896e0749cb7b4189c"
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
