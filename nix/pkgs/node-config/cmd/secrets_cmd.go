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

func SecretCmd(ctx *cli.Context) error {
	//We should maybe offer the means to not update but generate yaml files instead that then can be updated... but id didn't want to manage a bunch of structs just for the yaml stuff...
	var err error

	err = AskPostgresPassword(ctx.Context, clientset)
	if err != nil { //failed to create postgress password... find out why
		return err
	}
	err = AskKeyCloakPassword(ctx.Context, clientset)
	if err != nil { //failed to create keycloak password... find out why
		return err
	}
	err = PrepareArgoCd(ctx.Context, clientset)

	return err
}

func AskPostgresPassword(ctx context.Context, client kubernetes.Interface) error {
	pwd, err := askPassword("postgres admin account")
	if err != nil {
		return err
	}
	return CreateOrUpdateSecret(ctx, client, "default", "postgres-users", map[string]string{
		"postgres.password": pwd,
	}, map[string]string{})
}

func AskKeyCloakPassword(ctx context.Context, client kubernetes.Interface) error {
	pwd, err := askPassword("keycloak admin account")
	if err != nil {
		return err
	}
	return CreateOrUpdateSecret(ctx, client, "default", "keycloak-builtin-admin", map[string]string{
		"username": "admin",
		"password": pwd,
	}, map[string]string{})
}

func PrepareArgoCd(ctx context.Context, client kubernetes.Interface) error {
	//first we setup the repo secret
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Please enter or update your Teadal Node repo URL: ")
	argoURL, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	argoURL = strings.TrimSuffix(argoURL, "\n")

	fmt.Print("Please enter the deployment token username (generated on Gitlab):")
	username, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	username = strings.TrimSuffix(username, "\n")

	token, err := askPassword("Please enter the deployment token (generated on Gitlab)")
	if err != nil {
		return err
	}

	//for this secret we dont use the sec.Type filed thus, we use the lower level func
	err = createArgoRepoFrom(ctx, argoURL, username, token, client)

	if err != nil {
		return fmt.Errorf("failed to create argo repo secret %+v", err)
	}

	pwd, err := askPassword("enter the admin password for argo")
	if err != nil {
		return err
	}

	err = createArgoSecretFromPassword(ctx, client, pwd)

	if err != nil {
		return fmt.Errorf("failed to set argo account %+v", err)
	}
	return nil
}
