package main

import (
	"fmt"
	"testing"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

type Args struct {
	S string
}

type Result struct {
	String string
	Int    int
	Args   *Args
}

func TestGetHeight(t *testing.T) {
	nc, err := NewDefaultConnection()
	assert.NoError(t, err)
	height, err := nc.Height()
	assert.NoError(t, err)
	fmt.Println("height: ", height)
}

func TestGetExtrinsic(t *testing.T) {
	nc, err := NewConnection("wss://rpc.polkadot.io")
	assert.NoError(t, err)

	h, err := nc.Height()
	assert.NoError(t, err)
	fmt.Println("height: ", h)

	e, err := nc.GetExtrinsic(8085393, 2)
	assert.NoError(t, err)
	fmt.Println("e: ")
	fmt.Println(e)
	fmt.Printf("signer: %#x\n", e.Signature.Signer.AsID)

}

func TestQueryStorageAt(t *testing.T) {
	//	nc, err := NewConnection("wss://polkadot.api.onfinality.io/public-ws")
	bob := "14E5nqKAp3oAJcmzgZhUD2RcptBeUBScxKHgJKU4HPNcKVf3"
	nc, err := NewDefaultConnection()
	assert.NoError(t, err)

	genesis, err := nc.GetGenesisHash()
	assert.NoError(t, err)

	// Transfers to Bob - query storage against Bob's address, and genesis block hash...
	results, err := nc.QueryStorageAt(bob, genesis)
	assert.NoError(t, err)

	fmt.Println("results: ", results)

}

func TestGetBlockByHash(t *testing.T) {
	nc, err := NewDefaultConnection()
	assert.NoError(t, err)
	//	blockHashStr := "0xf610dd6437d0075f7b505af8e77ea67643e1dd0bab072945fc4c62b16b096352"
	blockHashStr := "0xf610dd6437d0075f7b505af8e77ea67643e1dd0bab072945fc4c62b16b096352"
	blockHashBytes := types.MustHexDecodeString(blockHashStr)
	blockHash := types.NewHash(blockHashBytes)

	fmt.Println("blockHash: ", blockHash)

	block, err := nc.GetBlockByHashTimeout(blockHash)
	assert.NoError(t, err)
	fmt.Println("block:", block)

}

func TestHealth(t *testing.T) {
	nc, err := NewDefaultConnection()
	assert.NoError(t, err)

	health, err := nc.HealthReportTimeout(1)
	assert.NoError(t, err)

	fmt.Println("peers: ", health.Peers)
	fmt.Println("syncing: ", health.IsSyncing)
	fmt.Println("should ", health.ShouldHavePeers)

	state, err := nc.Api.RPC.System.NetworkState()
	assert.NoError(t, err)
	fmt.Println("state: ", state.PeerID)

}
