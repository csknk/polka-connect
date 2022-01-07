package main

import (
	"fmt"
	"log"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

func ExampleHashToBytes() {
	blockHashStr := "0xbd6d7bde65925fc679015ac8b2da8b8d24e92f4361d9ca604c1e39af09eccb37"
	hash, err := types.NewHashFromHexString(blockHashStr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%#x\n", HashToBytes(hash))

	// Output: 0xbd6d7bde65925fc679015ac8b2da8b8d24e92f4361d9ca604c1e39af09eccb37
}
