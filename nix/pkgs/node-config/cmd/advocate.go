package main

import (
	"context"
	"fmt"
	"strings"

	core "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)


func ConfigureAdvocateSecrets(ctx context.Context, client kubernetes.Interface,
	eth_rpc_address, wallet_key string,
	permissioned_chain bool) error {

	node_key, err := GenerateRandomStringURLSafe(16)
	if err != nil {
		return fmt.Errorf("failed to generate nodeKey %+v", err)
	}
	sec := &core.Secret{}
	sec.Name = "advocate-wallet"
	sec.Namespace = "trust-plane"
	sec.Type = "Opaque"

	if !strings.HasSuffix("\"", wallet_key) && !strings.HasPrefix("\"", wallet_key) {
		wallet_key = fmt.Sprintf("\"%s\"", wallet_key)
	}

	/*if !strings.HasPrefix("\"0x", wallet_key) {
		return fmt.Errorf("The wallet key should be start with 0x and be in hexerdecimal format.")
	}*/

	sec.StringData = map[string]string{
		"ADVOCATE_WALLET_PRIVATEKEY_FILE": wallet_key,
		"ADVOCATE_VM_KEY":                 node_key,
		"ADVOCATE_ETH_RPC_ADDRESS":        eth_rpc_address,
	}

	err = CreateOrUpdateSecretWithStruct(ctx, client, sec)
	if err != nil {
		return fmt.Errorf("failed to create advocate-wallet secrete %+v", err)
	}

	if permissioned_chain {
		err := CreateOrUpdateConfig(ctx, client, "trust-plane", "advocate-config", map[string]string{
			"ADVOCATE_ETH_POA":        "1",
			"ADVOCATE_ETH_GAS_PRICE":  "0",
			"ADVOCATE_INSECURE_HTTPS": "1",
		})
		if err != nil {
			return fmt.Errorf("failed to update config to use a permissioned chain %+v", err)
		}
	}

	return nil
}
