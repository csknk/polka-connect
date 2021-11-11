package main

import (
	"github.com/centrifuge/go-substrate-rpc-client/v3/types"
)

func (c *Connection) getMetadata(blockHash types.Hash) (*types.Metadata, error) {
	return c.Api.RPC.State.GetMetadata(blockHash)
}

func (c *Connection) getLatestMetadata() (*types.Metadata, error) {
	return c.Api.RPC.State.GetMetadataLatest()
}
