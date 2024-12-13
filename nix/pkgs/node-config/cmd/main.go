package main

import (
	"fmt"
	"os"
	"syscall"

	"github.com/urfave/cli/v2"
	"golang.org/x/term"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// kubeclient is availible as a global ...
var clientset kubernetes.Interface

func main() {
	app := &cli.App{
		Name:        "node.config",
		Description: "Teadal config tool, to manage a teadal node",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "kubeconfig",
				Value: "",
				Usage: "provide the path to the kubernetes config to use",
			},
			&cli.BoolFlag{
				Name:  "microk8s",
				Usage: "use micok8s config for kubernetes, ignores path setting.",
			},
		},
		Before: func(ctx *cli.Context) error {
			var config *rest.Config
			var err error
			if ctx.Bool("microk8s") {
				config, err = clientcmd.BuildConfigFromFlags("", "/var/snap/microk8s/current/credentials/client.config")
			} else if configPath := ctx.String("kubeconfig"); configPath != "" {
				config, err = clientcmd.BuildConfigFromFlags("", configPath)
			} else {
				// Use the current context in kubeconfig
				config, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
			}
			if err != nil {
				return err
			}
			fmt.Printf("using the following kubernetes deployment: %+v. Use Cntr+C if thats not right...\n\n", config.Host)
			// Create the clientset
			clientset, err = kubernetes.NewForConfig(config)
			if err != nil {
				return err
			}
			fmt.Printf("Welcome to the Teadal node config cli. This tool will ask you to provide the passwords for a newly created teadal node instance.\n\n")
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:   "basicnode-secrets",
				Usage:  "set/reset secrets for basic teadal node installation",
				Action: BasicNodeSecretCmd,
			},
			{
				Name:   "postgres-secrets",
				Usage:  "set/reset postgres secrets",
				Action: PostgresSecretCmd,
			},
			{
				Name:   "keycloak-secrets",
				Usage:  "set/reset keycloak secrets",
				Action: KeycloakSecretCmd,
			},
			{
				Name:   "advocate",
				Usage:  "configure advocate tool",
				Action: AdvocateCmd,
			},
			{
				Name:   "pv",
				Usage:  "generate pvs for your teadal node",
				Action: PvCmd,
				Subcommands: []*cli.Command{
                    {
                        Name:  "help",
                        Usage: "print command help",
                        Action: PrintHelp,
                        },
                    },
			},
		},
		After: func(ctx *cli.Context) error {
			fmt.Println("🎉 you are all set, happy TEADALing 🎉")
			return nil
		},
		Action: func(ctx *cli.Context) error {
				return cli.ShowAppHelp(ctx)
		},
	}
	if err := app.Run(os.Args); err != nil {
		panic(err)
	}

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
