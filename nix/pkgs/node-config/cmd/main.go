package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"syscall"
	"time"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	// Use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		panic(err.Error())
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("using the following kubernetes deployment: %+v. Use Cntr+C if thats not right...\n\n", config.Host)

	//TODO: we should probly. check if the config is the onc intend...

	mainCtx := context.Background()
	fmt.Printf("Welcome to the Teadal node config cli. This tool will ask you to provide the passwords for a newly created teadal node instance.\n\n")

	//We should maybe offer the means to not update but generate yaml files instead that then can be updated... but id didn't want to manage a bunch of structs just for the yaml stuff...

	err = AskPostgresPassword(mainCtx, clientset)
	if err != nil { //failed to create postgress password... find out why
		panic(err)
	}

	err = AskKeyCloakPassword(mainCtx, clientset)
	if err != nil { //failed to create keycloak password... find out why
		panic(err)
	}

	err = PrepareArgoCd(mainCtx, clientset)
	if err != nil {
		panic(err)
	}

	fmt.Println("🎉 you are all set, happy TEADALing 🎉")
}

// ask for password or panics
func askPassword(name string) (string, error) {
	fmt.Printf("Please provide the password for %s:\n", name)

	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	if len(bytePassword) == 0 {
		//TODO; we could imbrace some default polices here..
		return "", fmt.Errorf("password to short")
	}
	return string(bytePassword), nil

}

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
	token, err := askPassword("enter the argocd deployment token (generated on Gitlab)")
	if err != nil {
		return err
	}

	//for this secret we dont use the sec.Type filed thus, we use the lower level func
	sec := &core.Secret{}
	sec.Name = "teadal.node-repo"
	sec.Namespace = "argocd"
	sec.StringData = map[string]string{
		"type":     "git",
		"url":      argoURL,
		"username": "argocd",
		"password": token,
	}
	sec.Labels = map[string]string{
		"argocd.argoproj.io/secret-type": "repository",
	}

	err = CreateOrUpdateSecretWithStruct(ctx, client, sec)

	if err != nil {
		return fmt.Errorf("failed to create argo repo secret %+v", err)
	}

	pwd, err := askPassword("enter the admin password for argo")
	if err != nil {
		return err
	}

	currentTime := time.Now()
	formattedTime := currentTime.UTC().Format("2006-01-02T15:04:05Z")
	encodedTime := base64.StdEncoding.EncodeToString([]byte(formattedTime))

	hashedStr, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		panic(err.Error())
	}

	encodedPwd := base64.StdEncoding.EncodeToString(hashedStr)

	err = CreateOrUpdateSecret(ctx, client, "argocd", "argocd-secret", map[string]string{
		"admin.password":      encodedPwd,
		"admin.passwordMtime": encodedTime,
	}, map[string]string{
		"app.kubernetes.io/name":    "argocd-secret",
		"app.kubernetes.io/part-of": "argocd",
	})

	if err != nil {
		return fmt.Errorf("failed to set argo account %+v", err)
	}
	return nil
}
