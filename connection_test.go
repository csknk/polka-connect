package main

import (
	"fmt"
	"testing"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

//	nc, err := NewConnection("wss://rpc.polkadot.io")
//	nc, err := NewConnection("http://192.168.0.164:9933")
const Endpoint = "http://localhost:9934"
const ID = "0xf64f2fa5bee8d59dcc2038e1ccbf6fc1b26e72ed2037c8e546ab08409e6d172e"

type Args struct {
	S string
}

type Result struct {
	String string
	Int    int
	Args   *Args
}

func TestGetBlockHashes(t *testing.T) {
	nc, err := NewConnection(Endpoint)
	assert.NoError(t, err)
	hashes, err := nc.ChangedBlockHashes(ID, 9300000)
	assert.NoError(t, err)
	for _, hash := range hashes {
		fmt.Printf("%#x\n", hash)
	}
}

func TestGetHeight(t *testing.T) {
	nc, err := NewConnection(Endpoint)
	assert.NoError(t, err)
	height, err := nc.Height()
	assert.NoError(t, err)
	fmt.Println("height: ", height)
}

func TestNode(t *testing.T) {
	nc, err := NewConnection(Endpoint)
	assert.NoError(t, err)
	roles := []string{}
	err = nc.Api.Client.Call(&roles, "system_nodeRoles")
	assert.NoError(t, err)
	fmt.Println("roles: ", roles)
}

func TestGetBlock(t *testing.T) {
	nc, err := NewConnection(Endpoint)
	assert.NoError(t, err)

	cases := []struct {
		BlockHash string
		Height    uint64
	}{
		{"0x7ffa315b3bcc4e772d762fe7446091f36abeaec73f87497813417efd8423dd42", 9201611},
		{"0x36d7b7e6882ff5ea2737d0b8a5975467c8a672dda457494a8dae3f53afdfd913", 9201612},
		{"0xb52b7255ad79b7a60f4ac4e4b4ddd9d5c0acf1775faea39e4b6184e02b56ed33", 9201613},
		{"0x451c6aec18d874591e18982e0a34c1501eb5ba469dbaaf7f1b812d0ad9dcf7ce", 9317311},
		{"0x66a5f943b8f34ddd0693c982c40b2944355cb26fd4de86265517de2848c79b1a", 9317312},
		{"0x49fe4dc3c0ef7c3a713369b71844fbfc07b58e01f5897ff5ec49ec1ac5dae523", 9317313},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("%d, %s", tc.Height, tc.BlockHash), func(t *testing.T) {
			blockHash, err := types.NewHashFromHexString(tc.BlockHash)
			assert.NoError(t, err)
			fmt.Printf("Test block hash %s, height %d\n", tc.BlockHash, tc.Height)

			block, err := nc.GetBlock(blockHash)
			assert.NoError(t, err)

			assert.Equal(t, types.BlockNumber(tc.Height), block.Block.Header.Number)
		})
	}
}

func TestGetBlockSingle(t *testing.T) {
	nc, err := NewConnection(Endpoint)
	assert.NoError(t, err)
	blockHashStr := "0x36d7b7e6882ff5ea2737d0b8a5975467c8a672dda457494a8dae3f53afdfd913"
	//	blockHashStr := "0x7ffa315b3bcc4e772d762fe7446091f36abeaec73f87497813417efd8423dd42"
	var blockHeight uint64 = 9201612

	blockHash, err := types.NewHashFromHexString(blockHashStr)
	assert.NoError(t, err)
	fmt.Printf("Test block hash %s, height %d\n", blockHash, blockHeight)

	block, err := nc.GetBlock(blockHash)
	assert.NoError(t, err)

	assert.Equal(t, types.BlockNumber(blockHeight), block.Block.Header.Number)
	fmt.Println(block)

}

func TestGetExtrinsic(t *testing.T) {
	nc, err := NewConnection(Endpoint)
	assert.NoError(t, err)

	h, err := nc.Height()
	assert.NoError(t, err)
	fmt.Println("height: ", h)

	//	e, err := nc.GetExtrinsic(8085393, 2)
	e, err := nc.GetExtrinsic(9292238, 2)
	assert.NoError(t, err)
	fmt.Println("e: ")
	fmt.Println(e)
	fmt.Printf("signer: %#x\n", e.Signature.Signer.AsID)

}

func TestQueryStorageAt(t *testing.T) {
	//	nc, err := NewConnection("wss://polkadot.api.onfinality.io/public-ws")
	//	bob := "14E5nqKAp3oAJcmzgZhUD2RcptBeUBScxKHgJKU4HPNcKVf3"
	bob := "5HdfAETZTH5jTeP9rSCsqfF9kqRAZUQRL9ofj8wZBq676ua4"
	//	bob := "5GXCq8BcEzNrmqQN5avx3wARnMBcRyMJzJ5BeT9WrDukpErh"
	nc, err := NewConnection(Endpoint)
	assert.NoError(t, err)

	genesis, err := nc.GetGenesisHash()
	assert.NoError(t, err)

	// Transfers to Bob - query storage against Bob's address, and genesis block hash...
	results, err := nc.QueryStorageAt(bob, genesis)
	assert.NoError(t, err)

	fmt.Println("results: ", results)

}

func TestGetBlockByHash(t *testing.T) {
	nc, err := NewConnection(Endpoint)
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
	nc, err := NewConnection(Endpoint)
	assert.NoError(t, err)

	health, err := nc.HealthReportTimeout(1)
	assert.NoError(t, err)

	fmt.Println("peers: ", health.Peers)
	fmt.Println("syncing: ", health.IsSyncing)
	fmt.Println("should ", health.ShouldHavePeers)

	//	state, err := nc.Api.RPC.System.NetworkState()
	//	assert.NoError(t, err)
	//	fmt.Println("state: ", state.PeerID)

}
