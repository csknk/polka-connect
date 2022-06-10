package main

import (
	"bytes"
	"fmt"

	"github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

func (c *Connection) getMetadata(blockHash types.Hash) (*types.Metadata, error) {
	return c.Api.RPC.State.GetMetadata(blockHash)
}

func (c *Connection) getLatestMetadata() (*types.Metadata, error) {
	return c.Api.RPC.State.GetMetadataLatest()
}

func (c *Connection) getEvents(meta *types.Metadata) {
	for _, mod := range meta.AsMetadataV14.Pallets {
		if !mod.HasEvents {
			continue
		}
		buf := []byte{}
		b := bytes.NewReader(buf)
		decoder := scale.NewDecoder(b)
		e := mod.Events.Type.Decode(*decoder)
		fmt.Println("e:", e)
		fmt.Println("b:", b)
	}
}
