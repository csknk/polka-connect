package main

import (
	"fmt"
	"math/big"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

// GetStorageHistoryForID runs a storage query for the provided identity (public key represented as a hexadecimal
// string). Returns a slice of StorageChangeSet objects which comprise block hash and a slice of key-value
// representation of changes for this account in the given block.
// NOTE: Must be run against an ARCHIVAL node - a full node is insufficient since it does not retain a full
// block history.
func (c *Connection) GetStorageHistoryForID(ID string, checkpoint uint64) (changeData []types.StorageChangeSet, err error) {
	account, err := types.HexDecodeString(ID)
	if err != nil {
		return
	}

	meta, err := c.Api.RPC.State.GetMetadataLatest()
	if err != nil {
		err = fmt.Errorf("can't get meta for api: %w", err)
		return
	}

	storageKey, err := types.CreateStorageKey(meta, "System", "Account", account, nil)
	if err != nil {
		err = fmt.Errorf("can't create storage key to query balance for account %s: %w", ID, err)
		return
	}

	startBlockHash, err := c.Api.RPC.Chain.GetBlockHash(checkpoint)
	if err != nil {
		err = fmt.Errorf("GetBlockHash failed for account %s: %w", ID, err)
		return
	}

	changeData, err = c.Api.RPC.State.QueryStorageLatest([]types.StorageKey{storageKey}, startBlockHash)
	if err != nil {
		err = fmt.Errorf("QueryStorageLatest failed for account %s: %w", ID, err)
	}
	return
}

// ChangedBlockHashes returns a slice of block hashes for blocks in which the System.Account balance of the
// specified ID changed.
func (c *Connection) ChangedBlockHashes(ID string, checkpoint uint64) (blockHashes [][]byte, err error) {
	changes, err := c.GetStorageHistoryForID(ID, checkpoint)
	if err != nil {
		return
	}
	for _, change := range changes {
		blockHashes = append(blockHashes, change.Block[:])
	}
	return
}

func (c *Connection) ChangedBlockHashesUnique(ID string, checkpoint uint64) (blockHashes map[[32]byte]bool, err error) {
	changes, err := c.GetStorageHistoryForID(ID, checkpoint)
	if err != nil {
		return
	}
	blockHashes = make(map[[32]byte]bool)
	for _, change := range changes {
		key := (*[32]byte)(change.Block[:])
		blockHashes[*key] = true
	}
	return
}

// DecodeAccountInfo decodes raw bytes into an AccountInfo struct.
func (c *Connection) DecodeAccountInfo(rawBytes []byte) (accountInfo AccountInfo) {
	types.DecodeFromBytes(rawBytes, &accountInfo)
	return
}

type ChangeData struct {
	blockHash         []byte
	publicKey         []byte
	ID                string
	amountAtThisBlock big.Int //types.U128
}

// GetChangeData --
func (c *Connection) GetChangeData(ID string, checkpoint uint64) (changeDataCollection []ChangeData, err error) {
	changes, err := c.GetStorageHistoryForID(ID, checkpoint)
	if err != nil {
		return
	}

	for _, change := range changes {
		amount := c.DecodeAccountInfo(change.Changes[0].StorageData).Data.Free
		res := ChangeData{
			blockHash:         change.Block[:],
			ID:                ID,
			amountAtThisBlock: *amount.Int,
		}
		changeDataCollection = append(changeDataCollection, res)
	}
	return
}
