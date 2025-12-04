package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all deployments",
	Long:  `Display a list of all deployments from YAML files in the fdps folder.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runList()
	},
}

type DeploymentConfig struct {
	FdpDescriptor struct {
		Name  string `yaml:"name"`
		Image struct {
			Name string `yaml:"name"`
		} `yaml:"image"`
	} `yaml:"fdp-descriptor"`
}

func runList() error {
	fdpsDir := filepath.Join(Config.Paths.Node, Config.Paths.ArgoCD, "fdps")
	fmt.Println(fdpsDir)

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

	fmt.Println("Deployed FDPs:")
	for _, d := range deployments {
		fmt.Printf("- Name: %s, Image: %s\n", d.FdpDescriptor.Name, d.FdpDescriptor.Image.Name)
	}

	return nil
}
