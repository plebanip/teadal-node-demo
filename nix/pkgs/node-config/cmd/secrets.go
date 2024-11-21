package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

func createArgoRepoFrom(ctx context.Context, argoURL string, username string, token string, client kubernetes.Interface) error {
	sec := &core.Secret{}
	sec.Name = "teadal.node-repo"
	sec.Namespace = "argocd"
	sec.StringData = map[string]string{
		"type":     "git",
		"url":      argoURL,
		"username": username,
		"password": token,
	}
	sec.Labels = map[string]string{
		"argocd.argoproj.io/secret-type": "repository",
	}

	return CreateOrUpdateSecretWithStruct(ctx, client, sec)
}

func createArgoSecretFromPassword(ctx context.Context, client kubernetes.Interface, pwd string) error {
	currentTime := time.Now()
	formattedTime := currentTime.UTC().Format("2006-01-02T15:04:05Z")

	bcryptpwd, err := bcrypt.GenerateFromPassword([]byte(pwd), 0)
	if err != nil {
		return fmt.Errorf("failed to cytpy password, cause %+v", err)
	}

	sec := &core.Secret{}
	sec.Name = "argocd-secret"
	sec.Namespace = "argocd"
	sec.Data = map[string][]byte{
		"admin.password": bcryptpwd,
	}
	sec.StringData = map[string]string{
		"admin.passwordMtime": formattedTime,
	}
	sec.Labels = map[string]string{
		"app.kubernetes.io/name":    "argocd-secret",
		"app.kubernetes.io/part-of": "argocd",
	}

	return CreateOrUpdateSecretWithStruct(ctx, client, sec)

}

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

	if !strings.HasPrefix("\"0x", wallet_key) {
		return fmt.Errorf("The wallet key should be start with 0x and be in hexerdecimal format.")
	}

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
