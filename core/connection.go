package core

import (
	"bytes"
	"fmt"
	"math/big"
	"time"

	"github.com/centrifuge/go-substrate-rpc-client/config"
	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/decred/base58"
	"golang.org/x/crypto/blake2b"
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

// NewDefaultConnection provides a GSRPC API connection to a Substrate node using the default address.
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

func (c *Connection) GetBalance(id string) (types.U128, types.U32, error) {
	account, err := types.HexDecodeString(id)
	zero := types.NewU128(*big.NewInt(0))
	zeroNonce := types.NewU32(uint32(0))

	if err != nil {
		return zero, zeroNonce, err
	}

	meta, err := c.Api.RPC.State.GetMetadataLatest()
	if err != nil {
		return zero, zeroNonce, fmt.Errorf("can't get meta for api: %w", err)
	}

	key, err := types.CreateStorageKey(meta, "System", "Account", account, nil)
	if err != nil {
		return zero, zeroNonce, err
	}

	var accountInfo AccountInfo
	ok, err := c.Api.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil || !ok {
		return zero, zeroNonce, err
	}

	num := accountInfo.Data.Free
	nonce := accountInfo.Nonce
	return num, nonce, nil
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
		err = fmt.Errorf("QueryStorageLatest error, startBlockHash %#x: %w", startBlockHash, err)
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

// PublicKeyFromAddress returns public key bytes from supplied SS58 address. Checks validity of address checksum.
func PublicKeyFromAddress(address string) (publicKey []byte, err error) {
	addressChecksum, publicKey, networkByte, err := addressComponents(address)
	if err != nil {
		err = fmt.Errorf("error splitting address components for %s: %w", address, err)
		return
	}

	computedChecksum, err := Checksum(publicKey, networkByte)
	if err != nil {
		err = fmt.Errorf("error getting checksum for address %s: %w", address, err)
		return
	}

	if !bytes.Equal(computedChecksum[:2], addressChecksum) {
		err = fmt.Errorf("invalid checksum for address %s", address)
		return
	}
	return publicKey, nil
}

// addressComponents breaks a valid substrate address into it's checksum, public key bytes and network byte.
// Address validity is not checked here - the returned data is used to validate the address by the caller.
func addressComponents(address string) (addressChecksum, publicKey []byte, networkByte uint8, err error) {
	if len(address) == 0 {
		return nil, nil, 0, fmt.Errorf("addressComponents received empty address parameter")
	}
	addressBytes := base58.Decode(address)
	// Minimum len = 1 (network byte) + 32 (public key) + 2 (minimum checksum) = 35
	// NOTE: It is possible that the checksum length may be greater than 2.
	if len(addressBytes) < 35 {
		return nil, nil, 0, fmt.Errorf("addressComponents() length of addressBytes (%v) is too short for address %s", len(addressBytes), address)
	}
	addressChecksum = addressBytes[len(addressBytes)-2:]
	publicKey = addressBytes[1 : len(addressBytes)-2]
	networkByte = addressBytes[0]
	len := len(publicKey)
	if len != 32 {
		err = fmt.Errorf("SS58 address %s yielded wrong length (%d) public key", address, len)
		return nil, nil, 0, err
	}
	return
}

// Checksum computes the checksum hash for a Polkadot/Substrate public key. The value of networkByte for
// Polkadot should be 0. Note that the bytes to hash are prepended with a set of magic bytes `ss58Prefix`.
func Checksum(publicKey []byte, networkByte uint8) ([]byte, error) {
	base := append([]byte{networkByte}, publicKey...)
	hasher, err := blake2b.New(64, nil)
	if err != nil {
		return nil, err
	}
	ss58Prefix := "SS58PRE"
	_, err = hasher.Write(append([]byte(ss58Prefix), base...))
	if err != nil {
		return nil, err
	}
	return hasher.Sum(nil), nil
}
