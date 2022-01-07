PolkaDot RPC Client
===================

Interface with a specified Polkadot/Substrate node.

```bash
curl -H "Content-Type: application/json" -d '{"id":1, "jsonrpc":"2.0", "method": "chain_getBlockHash"}' http://localhost:9933/
# Output: 
{"jsonrpc":"2.0","result":"0xac02e9d5ed9be9b4104c7888a1e3bd81b4ac291aa5cc610216663fffee009259","id":1}
```

Equivalent to:

```go
api, err := gsrpc.NewSubstrateAPI(config.Default().RPCURL)
if err != nil {
	log.Fatal(err)
}
hash, err := api.RPC.Chain.GetBlockHashLatest()
if err != nil {
	log.Fatal(err)
}
fmt.Println(hash.Hex())
```

Extrinsic Hash
--------------
Get the hash that identifies an extrinsic by passing the signed extrinsic to `types.GetHash()`:

```go
h, _ := types.GetHash(extrinsic)
fmt.Printf("%#x\n", h)
```

Submit Extrinsic for Inclusion in the L1 Chain
----------------------------------------------
The `author_submitExtrinsic` method just returns the extrinsic hash - it has no knowledge of validation or chain inclusion status.

```go
hash, err := c.Api.RPC.Author.SubmitExtrinsic(extrinsic)
fmt.Printf("Transfer sent with hash %#x\n", hash) // sent, but included?
```

See: https://pkg.go.dev/github.com/centrifuge/go-substrate-rpc-client?utm_source=godoc#example-package-MakeASimpleTransfer
See: https://pkg.go.dev/github.com/centrifuge/go-substrate-rpc-client?utm_source=godoc#hdr-Signing_extrinsics

Coding Notes
------------
Convert `types.Hash` to a `[]byte` type:

```go
// HashToBytes converts types.Hash to a []byte.
// rawBytes := hashType[:]
func HashToBytes(hash types.Hash) []byte {
	return hash[:]
}
```
