package main

import (
	"fmt"
	"log"
)

type State struct {
	startingBlock uint64 `json:"starting_block"`
	currentBlock  uint64 `json:"current_block"`
	highestBlock  uint64 `json:"highest_block"`
}

func NetworkState() (state State, err error) {
	nc, err := NewDefaultConnection()
	if err != nil {
		log.Fatal(err)
	}

	health, err := nc.Api.RPC.System.Health()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("nPeers:", health.Peers) // All OK here: {1 false false}

	//	state, err = nc.Api.RPC.System.NetworkState()
	//	peers, err = nc.SyncState()
	//	if err != nil {
	//		log.Fatal(err) // Method not found
	//	}
	//	fmt.Println("state: ", state)

	return
}

// NetworkState retrieves the current state of the network
//func (c *System) NetworkState() (types.NetworkState, error) {
func (c *Connection) GetPeers() ([]string, error) {
	var peers []string
	err := c.Api.Client.Call(&peers, "system_localListenAddresses")
	return peers, err
}

func (c *Connection) listening() error {

	l, err := c.Api.RPC.System.Properties()
	if err != nil {
		return err
	}
	fmt.Println("l: ", l)
	return nil

}
