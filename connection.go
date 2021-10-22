package main

import (
	"fmt"
	"math/big"

	"github.com/centrifuge/go-substrate-rpc-client/config"
	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v3"
	"github.com/centrifuge/go-substrate-rpc-client/v3/signature"
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

func NewConnection() (*Connection, error) {
	c := Connection{}
	var err error
	c.Api, err = gsrpc.NewSubstrateAPI(config.Default().RPCURL)
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
		return zero, err
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

func (c *Connection) Transfer(from, to string, amount uint64) error {

	meta, err := c.Api.RPC.State.GetMetadataLatest()
	if err != nil {
		return err
	}

	recipient, err := types.NewMultiAddressFromHexAccountID(to)
	if err != nil {
		return err
	}

	call, err := types.NewCall(meta, "Balances.transfer", recipient, types.NewUCompactFromUInt(amount))
	if err != nil {
		return fmt.Errorf("broblem making new call: %w", err)
	}
	extrinsic := types.NewExtrinsic(call)

	genesisHash, err := c.Api.RPC.Chain.GetBlockHash(0)
	if err != nil {
		return err
	}

	runtimeVersion, err := c.Api.RPC.State.GetRuntimeVersionLatest()
	if err != nil {
		return fmt.Errorf("problem getting latest version of runtime: %w", err)
	}

	// -----------
	fr, ok := signature.LoadKeyringPairFromEnv()
	if !ok {
		fr = signature.TestKeyringPairAlice
	}
	// -----------

	//	key, err := types.CreateStorageKey(meta, "System", "Account", signature.TestKeyringPairAlice.PublicKey)
	key, err := types.CreateStorageKey(meta, "System", "Account", fr.PublicKey)
	if err != nil {
		return fmt.Errorf("problem creating storage key: %w", err)
	}

	var accountInfo types.AccountInfo

	ok, err = c.Api.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil || !ok {
		return fmt.Errorf("problem getting accountInfo: %w", err)
	}

	nonce := uint32(accountInfo.Nonce)

	o := types.SignatureOptions{
		BlockHash:   genesisHash,
		Era:         types.ExtrinsicEra{IsMortalEra: false},
		GenesisHash: genesisHash,
		Nonce:       types.NewUCompactFromUInt(uint64(nonce)),
		SpecVersion: runtimeVersion.SpecVersion,
		Tip:         types.NewUCompactFromUInt(0),
	}

	// Sign transaction
	if err := extrinsic.Sign(signature.TestKeyringPairAlice, o); err != nil {
		return fmt.Errorf("problem signing: %w", err)
	}

	hash, err := c.Api.RPC.Author.SubmitExtrinsic(extrinsic)
	if err != nil {
		return fmt.Errorf("submit extrinsic: %w", err)
	}
	fmt.Printf("Transfer sent with hash %#x\n", hash)

	return nil
}
