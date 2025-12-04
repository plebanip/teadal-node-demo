package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "fdp-manager",
	Short: "A CLI application for managing deployments",
	Long:  `A command-line application that provides list, deploy, and undeploy commands.`,
}

func Execute() error {
	if err := LoadConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Config loaded successfully.\n")
	fmt.Printf("Node general path: %s\n", Config.Paths.Node)
	fmt.Printf("ArgoCD conf path: %s\n", Config.Paths.ArgoCD)
	fmt.Printf("Micork8s conf path: %s\n", Config.Paths.MicroK8S)

	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(undeployCmd)
}
