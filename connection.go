package main

import (
	"fmt"
	"math/big"
	"time"

	"github.com/centrifuge/go-substrate-rpc-client/config"
	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

const (
	// Timeout for Polkadot node healthcheck ping
	Timeout                  = 10 * time.Second
	MainnetGenesisHashString = "0x91b171bb158e2d3848fa23a9f1c25182fb8e20313b2c1eb49219da7a70ce90c3"
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

// HealthReportTimeout returns nil if the healthcheck occurs within the provided timeout (seconds), otherwise
// returns an appropriate error.
func (c *Connection) HealthReportTimeout(timeout int) (health *types.Health, err error) {
	errorCh := make(chan error, 1)
	resultCh := make(chan *types.Health, 1)
	go func() {
		//		time.Sleep(2 * time.Second)
		health, err := c.Api.RPC.System.Health()
		if err != nil {
			errorCh <- err
			return
		}
		resultCh <- &health
	}()

	select {
	case err = <-errorCh:
		return
	case health = <-resultCh:
		return
	case <-time.After(time.Duration(timeout) * time.Second):
		err = fmt.Errorf("healthcheck of Polkadot node timeout exceeded")
		return
	}
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

func (c *Connection) Height() (uint64, error) {
	header, _ := c.Api.RPC.Chain.GetHeaderLatest()
	return uint64(header.Number), nil
}

// GetGenesisHash returns the Genesis Hash of the connected network
func (c *Connection) GetGenesisHash() (genesisHash types.Hash, err error) {
	return c.Api.RPC.Chain.GetBlockHash(1)
}

// GetExtrinsic returns a signed extrinsic given a block height and index
func (c *Connection) GetExtrinsic(height uint64, index uint64) (extrinsic types.Extrinsic, err error) {
	blockHash, err := c.Api.RPC.Chain.GetBlockHash(height)
	if err != nil {
		return
	}

	block, err := c.Api.RPC.Chain.GetBlock(blockHash)
	if err != nil {
		return
	}

	extrinsic = block.Block.Extrinsics[index]

	return extrinsic, nil
}

func (c *Connection) QueryStorageAt(address string, startBlockHash types.Hash) (storage []types.StorageChangeSet, err error) {
	accountID, err := PublicKeyFromAddress(address)
	if err != nil {
		return nil, err
	}

	meta, err := c.Api.RPC.State.GetMetadataLatest()
	if err != nil {
		return
	}

	// Build a key that will be used to fetch account balance for the specified account
	key, err := types.CreateStorageKey(meta, "System", "Account", accountID, nil)
	if err != nil {
		err = fmt.Errorf("failed to create storage key: %w", err)
		return
	}

	storage, err = c.Api.RPC.State.QueryStorageLatest([]types.StorageKey{key}, startBlockHash)
	if err != nil {
		fmt.Println("storage return")
		fmt.Println("err:", err)

		return
	}
	//	fmt.Println("storage: ", storage)
	fmt.Println("len of storage: ", len(storage))

	for i, res := range storage {
		fmt.Println("i: ", i)
		fmt.Println("block: ", res.Block.Hex())
		//		fmt.Println("changes: ", res.Changes)
		for j, el := range res.Changes {
			fmt.Printf("change %v storage key: %v\n", j, el.StorageKey)
			fmt.Printf("storage id from key: %#x\n", el.StorageKey[len(el.StorageKey)-32:])

			fmt.Printf("change %v has storage data: %v\n", j, el.HasStorageData)
			fmt.Println("storage data: ", el.StorageData)
			var acc types.AccountInfo
			if err = types.DecodeFromBytes(el.StorageData, &acc); err != nil {
				panic(err)
			}

			fmt.Println("acc: ", acc)

			//			buf := bytes.Buffer{}
			//			enc := scale.NewEncoder(&buf)
			//			data := el.StorageData.Encode(*enc)
			//			fmt.Printf("change %v storage data: %v\n", j, data)
		}

	}
	return
}

// GetBlockByHeight gets the block at the required height. A timeout is applied.
func (c *Connection) GetBlockByHash(blockHash types.Hash) (block *types.SignedBlock, err error) {
	block, err = c.Api.RPC.Chain.GetBlock(blockHash)
	return
}

// GetBlockByHeight gets the block at the required height. A timeout is applied.
func (c *Connection) GetBlockByHashTimeout(blockHash types.Hash) (block *types.SignedBlock, err error) {
	errorCh := make(chan error, 1)
	blockCh := make(chan *types.SignedBlock, 1)
	go func() {
		blk, err := c.Api.RPC.Chain.GetBlock(blockHash)
		if err != nil {
			errorCh <- err
			return
		}
		blockCh <- blk
	}()

	select {
	case err = <-errorCh:
	case block = <-blockCh:
	case <-time.After(Timeout):
		err = fmt.Errorf("failed to get Polkadot block")
	}
	return
}
