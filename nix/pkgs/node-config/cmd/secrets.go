package main

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CreateOrUpdateSecretWithStruct creates or updates a Kubernetes secret with the provided struct.
// It first tries to get the secret, if it exists it updates it, otherwise it creates a new one.
// It returns an error if the creation or update operation fails.
func CreateOrUpdateSecretWithStruct(ctx context.Context, client kubernetes.Interface, sec *core.Secret) error {
	secrets_client := client.CoreV1().Secrets(sec.Namespace)
	_sec, err := secrets_client.Get(ctx, sec.Name, v1.GetOptions{})

	if err != nil { //for now assuem it didn't exsist and just create it..
		//Not perfect, we should porbly. use a template?

		_, err := secrets_client.Create(ctx, sec, v1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create %s password. You are on your own. %+v", sec.Name, err)
		}
	} else {
		//TODO: implemet a merge instead..
		_sec.StringData = sec.StringData
		_, err := secrets_client.Update(context.Background(), _sec, v1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update %s password. You are on your own. %+v", sec.Name, err)
		}
	}
	return nil
}

// CreateOrUpdateSecret creates or updates a Kubernetes secret.
// It takes in parameters for the context, client, namespace, name, data, and labels.
// It constructs a secret struct from the provided parameters and then calls CreateOrUpdateSecretWithStruct.
// It returns an error if the creation or update operation fails.
func CreateOrUpdateSecret(ctx context.Context, client kubernetes.Interface,
	namespace, name string, data, labels map[string]string) error {
	sec := &core.Secret{}
	sec.Name = name
	sec.Namespace = namespace
	sec.Type = core.SecretTypeOpaque
	sec.StringData = data
	return CreateOrUpdateSecretWithStruct(ctx, client, sec)
}

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
