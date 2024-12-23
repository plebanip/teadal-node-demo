package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
	"k8s.io/client-go/kubernetes"
)

func AdvocateCmd(ctx *cli.Context) error {
	//We should maybe offer the means to not update but generate yaml files instead that then can be updated... but id didn't want to manage a bunch of structs just for the yaml stuff...
	var err error

	err = PrepareAdvocate(ctx.Context, clientset)

	return err
}


func PrepareAdvocate(ctx context.Context, client kubernetes.Interface) error {
	fmt.Println("We will now set up adovcate.")

	wallet_key, err := askPassword("Provide your wallet private key")
	if err != nil {
		return err
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Please enter or update the eth rpc address you want to use:")
	eth_rpc_address, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	eth_rpc_address = strings.TrimSuffix(eth_rpc_address, "\n")

	fmt.Print("What type of etherium network are you using? [PoS=0,PoA=1,PoW=2]? ")
	type_string, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	var permissioned_mode bool = false // for now, might change later
	switch strings.TrimSuffix(type_string, "\n") {
	case "1":
		permissioned_mode = true
	}

	return ConfigureAdvocateSecrets(ctx, client, eth_rpc_address, wallet_key, permissioned_mode)
}
