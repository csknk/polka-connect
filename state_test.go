package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	westend = "wss://westend-rpc.polkadot.io"
	mainnet = "wss://rpc.polkadot.io"
)

func TestNetworkState(t *testing.T) {
	NetworkState()
}

func TestPeers(t *testing.T) {
	nc, err := NewConnection(westend)
	assert.NoError(t, err)

	peers, err := nc.GetPeers()
	assert.NoError(t, err)

	for _, peer := range peers {
		fmt.Println("peer: ", peer)
	}

}
