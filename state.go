package main

import (
	"fmt"
	"log"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

func NetworkState() (state types.NetworkState, err error) {
	nc, err := NewDefaultConnection()
	if err != nil {
		log.Fatal(err)
	}

	health, _ := nc.Api.RPC.System.Health()
	fmt.Println("health:", health)

	//	state, err = nc.Api.RPC.System.NetworkState()
	var h types.Health
	healthCall := nc.Api.Client.Call(&h, "system_health")
	fmt.Println("healthCall: ", healthCall)

	//	health, err := nc.Api.RPC.System.Health()
	//	fmt.Println("health: ", health)

	return
}
