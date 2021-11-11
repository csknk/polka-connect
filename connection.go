package main

import (
	"fmt"
	"math/big"

	"github.com/centrifuge/go-substrate-rpc-client/config"
	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v3"
	"github.com/centrifuge/go-substrate-rpc-client/v3/types"
)

// See: https://github.com/centrifuge/go-substrate-rpc-client/issues/154#issuecomment-850351285
type AccountInfo struct {
	Nonce       types.U32
	Consumers   types.U32
	Providers   types.U32
	Sufficients types.U32
	Data        struct {
		Free       types.U128
		Reserved   types.U128
		MiscFrozen types.U128
		FreeFrozen types.U128
	}
}

type Connection struct {
	Api *gsrpc.SubstrateAPI
}

func NewDefaultConnection() (*Connection, error) {
	return NewConnection("")
}

func NewConnection(endpoint string) (*Connection, error) {
	cfg := config.Default().RPCURL

	if endpoint != "" {
		cfg = endpoint
	}
	c := Connection{}
	var err error
	c.Api, err = gsrpc.NewSubstrateAPI(cfg)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *Connection) GetLatestBlockHash() (*types.Hash, error) {
	hash, err := c.Api.RPC.Chain.GetBlockHashLatest()
	if err != nil {
		return nil, err
	}
	return &hash, nil
}

func (c *Connection) GetBlock(hash types.Hash) (*types.SignedBlock, error) {
	block, err := c.Api.RPC.Chain.GetBlock(hash)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (c *Connection) HealthReport() {
	health, err := c.Api.RPC.System.Health()
	if err != nil {
		fmt.Println("Can't determine health")
	}
	fmt.Println("peers:", health.Peers)
	fmt.Println("is syncing:", health.IsSyncing)
	fmt.Println("should have peers: ", health.ShouldHavePeers)
}

func (c *Connection) GetBalance(id string) (types.U128, error) {
	account, err := types.HexDecodeString(id)
	zero := types.NewU128(*big.NewInt(0))
	if err != nil {
		return zero, err
	}

	meta, err := c.Api.RPC.State.GetMetadataLatest()
	if err != nil {
		return zero, fmt.Errorf("can't get meta for api: %w", err)
	}

	key, err := types.CreateStorageKey(meta, "System", "Account", account, nil)
	if err != nil {
		return zero, err
	}
	fmt.Printf("key: %#x\n", key)

	var accountInfo AccountInfo
	ok, err := c.Api.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil || !ok {
		return zero, err
	}

	num := accountInfo.Data.Free
	return num, nil
}

func (c *Connection) GetAddress(pubkey []byte) (types.Address, error) {

	address := types.NewAddressFromAccountID(pubkey)

	return address, nil
}
