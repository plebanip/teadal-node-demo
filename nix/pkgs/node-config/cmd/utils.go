package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

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
		_sec.Data = sec.Data
		_, err := secrets_client.Update(ctx, _sec, v1.UpdateOptions{})
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

func mergeMaps[K comparable, V any](original map[K]V, overwrite map[K]V) map[K]V {
	for key, value := range overwrite {
		original[key] = value
	}
	return original
}

func CreateOrUpdateConfig(ctx context.Context, client kubernetes.Interface,
	namesapce, name string, data map[string]string) error {
	conf, err := client.CoreV1().ConfigMaps(namesapce).Get(ctx, name, v1.GetOptions{})

	if err != nil { //for now assuem it didn't exsist and just create it..
		conf = &core.ConfigMap{
			ObjectMeta: v1.ObjectMeta{
				Name:      name,
				Namespace: namesapce,
			},
			Data: data,
		}
		_, err := client.CoreV1().ConfigMaps(namesapce).Create(ctx, conf, v1.CreateOptions{})
		return err
	} else { //lets update
		mergeMaps(conf.Data, data)
		_, err = client.CoreV1().ConfigMaps(namesapce).Update(ctx, conf, v1.UpdateOptions{})
		return err
	}
}

// GenerateRandomStringURLSafe returns a URL-safe, base64 encoded
// securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomStringURLSafe(n int) (string, error) {
	b, err := GenerateRandomBytes(n)
	return base64.URLEncoding.EncodeToString(b), err
}

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}
