package main

import (
	"context"
	"testing"

	core "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"
)

func TestCreateOrUpdateSecretWithStruct(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := testclient.NewSimpleClientset()

	_, err := client.CoreV1().Secrets("testing").Get(ctx, "test", v1.GetOptions{})
	if err == nil {
		t.Fail()
	}

	sec := &core.Secret{}
	sec.Name = "test"
	sec.Namespace = "testing"
	sec.StringData = map[string]string{
		"fake": "data",
	}
	sec.Labels = map[string]string{
		"with": "labels",
	}

	op := CreateOrUpdateSecretWithStruct(ctx, client, sec)

	if op != nil {
		t.Fail()
	}

	_, err = client.CoreV1().Secrets("testing").Get(ctx, "test", v1.GetOptions{})
	if err != nil {
		t.Fail()
	}
}

func TestCreateOrUpdateSecret(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := testclient.NewSimpleClientset()

	_, err := client.CoreV1().Secrets("testing").Get(ctx, "test", v1.GetOptions{})
	if err == nil {
		t.Fail()
	}

	op := CreateOrUpdateSecret(ctx, client, "testing", "test", map[string]string{
		"fake": "password",
	}, map[string]string{})

	if op != nil {
		t.Fail()
	}

	sec, err := client.CoreV1().Secrets("testing").Get(ctx, "test", v1.GetOptions{})
	if err != nil {
		t.Fail()
	}

	if sec.StringData["fake"] != "password" {
		t.Fail()
	}

	op = CreateOrUpdateSecret(ctx, client, "testing", "test", map[string]string{
		"fake": "changed",
	}, map[string]string{})

	if op != nil {
		t.Fail()
	}

	sec, err = client.CoreV1().Secrets("testing").Get(ctx, "test", v1.GetOptions{})
	if err != nil {
		t.Fail()
	}

	if sec.StringData["fake"] == "password" {
		t.Fail()
	}

	if sec.StringData["fake"] != "changed" {
		t.Fail()
	}
}
