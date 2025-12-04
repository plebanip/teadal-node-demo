package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var undeployCmd = &cobra.Command{
	Use:   "undeploy [name]",
	Short: "Undeploy a deployment by name",
	Long:  `Remove a deployment by specifying its name.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		return runUndeploy(name)
	},
}

func runUndeploy(fdpName string) error {
	if fdpName == "" {
		return fmt.Errorf("deployment name cannot be empty")
	}

	fdpsDir := filepath.Join(Config.Paths.Node, Config.Paths.ArgoCD, "fdps")

	// Check if fdps directory exists
	if _, err := os.Stat(fdpsDir); os.IsNotExist(err) {
		return fmt.Errorf("fdps directory not found")
	}

	// Read all files in fdps directory
	files, err := os.ReadDir(fdpsDir)
	if err != nil {
		return fmt.Errorf("failed to read fdps directory: %w", err)
	}

	deployments := []DeploymentConfig{}

	// Process each YAML file
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		ext := filepath.Ext(file.Name())
		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		filePath := filepath.Join(fdpsDir, file.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to read %s: %v\n", file.Name(), err)
			continue
		}

		var config DeploymentConfig
		if err := yaml.Unmarshal(data, &config); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to parse %s: %v\n", file.Name(), err)
			continue
		}

		deployments = append(deployments, config)
	}

	if len(deployments) == 0 {
		fmt.Println("No deployments found.")
		return nil
	}

	found := false
	for _, d := range deployments {
		if d.FdpDescriptor.Name == fdpName {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("FDP %q is not deployed", fdpName)
	}

	//starts the undeployment

	fmt.Printf("Undeploying: %s\n", fdpName)

	// remove the descriptor from fdps
	targetPath := filepath.Join(fdpsDir, fdpName+".yaml")

	if err := os.Remove(targetPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("FDP %q is not deployed (file %s does not exist)", fdpName, targetPath)
		}
		return fmt.Errorf("removing file %s: %w", targetPath, err)
	}

	// remove the folder under projects/pilot-services
	targetPath = filepath.Join(Config.Paths.Node, Config.Paths.ArgoCD, fdpName)
	if err := os.RemoveAll(targetPath); err != nil {
		return fmt.Errorf("removing folder %s: %w", targetPath, err)
	}
	fmt.Printf("Successfully removed resources for ArgoCD")

	// update the kustomization file
	targetDir := filepath.Join(Config.Paths.Node, Config.Paths.ArgoCD)
	if err := removeFromKustomizationResources(targetDir, fdpName); err != nil {
		return err
	}

	// remove the folder under deployment/pilot-services
	targetPath = filepath.Join(Config.Paths.Node, Config.Paths.MicroK8S, fdpName)
	if err := os.RemoveAll(targetPath); err != nil {
		return fmt.Errorf("removing folder %s: %w", targetPath, err)
	}
	fmt.Printf("Successfully removed resources for MicroK8S")

	fmt.Printf("Successfully undeployed: %s\n", fdpName)
	return nil
}
