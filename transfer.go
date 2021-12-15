package main

import (
	"fmt"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

func (c *Connection) Transfer(from, to string, amount uint64) error {

	extrinsic, err := c.NewExtrinsic(from, to, amount)
	if err != nil {
		return fmt.Errorf("error building new extrinsic: %w", err)
	}

	extrinsicString, err := types.EncodeToHexString(extrinsic)
	if err != nil {
		return err
	}

	fmt.Printf("extrinsic: %s\n", extrinsicString)

	subscription, err := c.Api.RPC.Author.SubmitAndWatchExtrinsic(*extrinsic)
	if err != nil {
		return fmt.Errorf("failure to submit extrinsic: %w", err)
	}
	defer subscription.Unsubscribe()
	for {
		status := <-subscription.Chan()
		if status.IsInBlock || status.IsFinalized {
			fmt.Printf("included in block %#x\n", status.AsInBlock)
			fmt.Println(status.MarshalJSON())

			break
		}
		fmt.Println("Waiting for extrinsic to be included in a block...")
	}

	return nil
}
