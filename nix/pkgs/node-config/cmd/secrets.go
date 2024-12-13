package main

import (
	"context"
	"fmt"
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

