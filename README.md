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

See: https://pkg.go.dev/github.com/centrifuge/go-substrate-rpc-client?utm_source=godoc#example-package-MakeASimpleTransfer

See: https://pkg.go.dev/github.com/centrifuge/go-substrate-rpc-client?utm_source=godoc#hdr-Signing_extrinsics

