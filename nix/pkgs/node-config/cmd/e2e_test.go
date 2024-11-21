package main

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	v1 "k8s.io/api/core/v1"

	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"sigs.k8s.io/e2e-framework/support/kind"
)

var (
	testenv env.Environment
)

const clusterName = "teadal.node.config.test"

// This test starts a real kube cluster using kind and runs the core functions against it... this will take tome time..
func TestMain(t *testing.M) {
	//IF the SKIP_E2E variable is set, we do not run these (for gitlab ci)
	if _, set := os.LookupEnv("SKIP_E2E"); set {
		os.Exit(t.Run())
	} else {
		testenv, _ = env.NewFromFlags()
		// pre-test setup of kind cluster
		testenv.Setup(
			envfuncs.CreateCluster(kind.NewProvider().WithOpts(kind.WithImage("kindest/node:v1.27.3")), clusterName),
			envfuncs.CreateNamespace("argocd"),
			envfuncs.CreateNamespace("trust-plane"),
		)
		// post-test teardown kind cluster
		testenv.Finish(
			envfuncs.DeleteNamespace("argocd"),
			envfuncs.DeleteNamespace("trust-plane"),
			envfuncs.DestroyCluster(clusterName),
		)
		os.Exit(testenv.Run(t))
	}

}

func Test_ArgoSecrets(t *testing.T) {
	if testenv == nil {
		t.Skip()
	}
	deploymentFeature := features.New("argocd/secret").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			client, err := kubernetes.NewForConfig(cfg.Client().RESTConfig())
			if err != nil {
				t.Fatalf("could not create test clinet %+v", err)
			}
			err = createArgoSecretFromPassword(ctx, client, "testpwd")
			if err != nil {
				t.Fatalf("could nto create argo secrets... %+v", err)
			}
			return ctx
		}).
		Assess("deployment creation", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			var sec v1.Secret
			if err := cfg.Client().Resources().Get(ctx, "argocd-secret", "argocd", &sec); err != nil {
				t.Fatal(err)
			}

			if err := bcrypt.CompareHashAndPassword(sec.Data["admin.password"], []byte("testpwd")); err != nil {
				t.Fatalf("expected testpwd but found %+v", err)
			}
			return context.WithValue(ctx, "argocd-secret", &sec)
		}).
		Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			dep := ctx.Value("argocd-secret").(*v1.Secret)
			if err := cfg.Client().Resources().Delete(ctx, dep); err != nil {
				t.Fatal(err)
			}
			return ctx
		}).Feature()

	testenv.Test(t, deploymentFeature)
}

func Test_AdvocateSecrets(t *testing.T) {
	if testenv == nil {
		t.Skip()
	}
	deploymentFeature := features.New("advocate/secret").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			client, err := kubernetes.NewForConfig(cfg.Client().RESTConfig())
			if err != nil {
				t.Fatalf("could not create test clinet %+v", err)
			}
			err = ConfigureAdvocateSecrets(ctx, client, "https://foo.bar:5435", "0x1251243145fasf1235124", true)
			if err != nil {
				t.Fatalf("could nto create advocate secrets... %+v", err)
			}
			return ctx
		}).
		Assess("deployment creation", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			var sec v1.Secret
			if err := cfg.Client().Resources().Get(ctx, "advocate-wallet", "trust-plane", &sec); err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, "\"0x1251243145fasf1235124\"", string(sec.Data["ADVOCATE_WALLET_PRIVATEKEY_FILE"]))
			assert.NotNil(t, sec.StringData["ADVOCATE_VM_KEY"])
			assert.Equal(t, "https://foo.bar:5435", string(sec.Data["ADVOCATE_ETH_RPC_ADDRESS"]))

			return context.WithValue(ctx, "advocate-wallet", &sec)
		}).
		Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			dep := ctx.Value("advocate-wallet").(*v1.Secret)
			if err := cfg.Client().Resources().Delete(ctx, dep); err != nil {
				t.Fatal(err)
			}
			return ctx
		}).Feature()

	testenv.Test(t, deploymentFeature)
}
