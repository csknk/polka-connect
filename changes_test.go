package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

//const (
//	blockHashStr    = "0x9f4b3c125a646033859053aa28101fe3d32c679d04ecca2fc5bfbf417401e671"
//	receiverPubKey  = "0x8eaf04151687736326c9fea17e25fc5287613693c912909cb226aa4794f26a48" // 14E5nqKAp3oAJcmzgZhUD2RcptBeUBScxKHgJKU4HPNcKVf3 Bob
//	receiverAddress = "14E5nqKAp3oAJcmzgZhUD2RcptBeUBScxKHgJKU4HPNcKVf3"
//)

func TestChangedBlockHashes(t *testing.T) {
	c, err := NewDefaultConnection()
	if err != nil {
		fmt.Println("No connection to node")
		assert.FailNow(t, err.Error())
	}

	changedBlocks, err := c.ChangedBlockHashesUnique(receiverPubKey, 1)
	assert.NoError(t, err)

	fmt.Printf("number of blocks: %d\n", len(changedBlocks))

	for block, _ := range changedBlocks {
		fmt.Printf("%#x\n", block)

	}
}
