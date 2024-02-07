package main

import (
	"context"
	"os"
	"testing"

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
		testenv, _ = env.NewFromFlags()
		// pre-test setup of kind cluster
		testenv.Setup(
			envfuncs.CreateCluster(kind.NewProvider().WithOpts(kind.WithImage("kindest/node:v1.27.3")), clusterName),
			envfuncs.CreateNamespace("argocd"),
		)
		// post-test teardown kind cluster
		testenv.Finish(
			envfuncs.DeleteNamespace("argocd"),
			envfuncs.DestroyCluster(clusterName),
		)
		os.Exit(t.Run())
	} else {
		os.Exit(testenv.Run(t))
	}

}

func Test_Secrets(t *testing.T) {
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
