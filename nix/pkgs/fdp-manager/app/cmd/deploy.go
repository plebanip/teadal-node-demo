package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var deployCmd = &cobra.Command{
	Use:   "deploy [descriptor-file]",
	Short: "Deploy with YAML descriptor file",
	Long:  `Deploy a new instance using the configuration specified in a YAML file.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		yamlDescriptorFile := args[0]
		return runDeploy(yamlDescriptorFile)
	},
}

func runDeploy(yamlDescriptorFile string) error {
	// Check if file exists
	if _, err := os.Stat(yamlDescriptorFile); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", yamlDescriptorFile)
	}

	// Check if file has .yaml or .yml extension
	ext := filepath.Ext(yamlDescriptorFile)
	if ext != ".yaml" && ext != ".yml" {
		return fmt.Errorf("file must be a YAML file (.yaml or .yml)")
	}

	// Read the YAML file
	data, err := os.ReadFile(yamlDescriptorFile)
	if err != nil {
		return fmt.Errorf("failed to read descriptor file: %w", err)
	}

	// Parse YAML
	var config FdpConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	fdpName := config.FdpDescriptor.Name
	if fdpName == "" {
		return fmt.Errorf("fdp-descriptor.name is required")
	}

	// Validate fdpName: no spaces and only valid folder name characters
	if strings.Contains(fdpName, " ") {
		return fmt.Errorf("fdp-descriptor.name cannot contain spaces: '%s'", fdpName)
	}

	// Check for invalid characters in folder names
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		if strings.Contains(fdpName, char) {
			return fmt.Errorf("fdp-descriptor.name contains invalid character '%s': '%s'", char, fdpName)
		}
	}

	// check if the deployment file and all the resources exists
	yamlDescriptorDir := filepath.Dir(yamlDescriptorFile)

	deploymentFile := filepath.Join(yamlDescriptorDir, config.FdpDescriptor.Deployment)
	data, err = os.ReadFile(deploymentFile)
	if err != nil {
		return fmt.Errorf("failed to read deployment file: %w", err)
	}

	for _, res := range config.FdpDescriptor.Resources {
		resourceFile := filepath.Join(yamlDescriptorDir, res)
		data, err = os.ReadFile(resourceFile)
		if err != nil {
			return fmt.Errorf("failed to read resource file: %w", err)
		}
	}

	// Check if directory already exists for ArgoCD
	pilotServicesArgoCDDir := filepath.Join(Config.Paths.Node, Config.Paths.ArgoCD)
	fdpArgoCDDir := filepath.Join(pilotServicesArgoCDDir, fdpName)

	if _, err := os.Stat(fdpArgoCDDir); err == nil {
		return fmt.Errorf("deployment '%s' already exists in pilot-services", fdpName)
	}

	// Check if directory already exists for Microk8s
	pilotServicesK8SDir := filepath.Join(Config.Paths.Node, Config.Paths.MicroK8S)
	fdpK8SDir := filepath.Join(pilotServicesK8SDir, fdpName)

	if _, err := os.Stat(fdpK8SDir); err == nil {
		return fmt.Errorf("deployment '%s' already exists in pilot-services", fdpName)
	}

	fmt.Printf("Deploying: %s\n", fdpName)

	//1. ArgoCD elements

	// Create pilot-services directory if it doesn't exist
	if err := os.MkdirAll(pilotServicesArgoCDDir, 0755); err != nil {
		return fmt.Errorf("failed to create pilot-services directory: %w", err)
	}

	// Create FDP-specific directory
	if err := os.MkdirAll(fdpArgoCDDir, 0755); err != nil {
		return fmt.Errorf("failed to create FDP directory: %w", err)
	}

	// Create app.yaml file
	appYamlContent := fmt.Sprintf(`- op: replace
  path: /metadata/name
  value: %s
- op: replace
  path: /spec/source/path
  value: deployment/pilot-services/%s
- op: replace
  path: /spec/project
  value: pilot-services
`, fdpName, fdpName)

	kustomizationArgoYamlContent := fmt.Sprintf(`apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- ../../base

patches:
- target:
    kind: Application
    name: app
  path: app.yaml
`)

	// add app.yaml for ArgoCD inside the pilot-service/fdp-descriptor.name folder

	appYamlPath := filepath.Join(fdpArgoCDDir, "app.yaml")
	if err := os.WriteFile(appYamlPath, []byte(appYamlContent), 0644); err != nil {
		return fmt.Errorf("failed to create app.yaml: %w", err)
	}
	fmt.Printf("Created: %s\n", appYamlPath)

	// add kustomization.yaml for ArgoCD inside the pilot-service/fdp-descriptor.name folder
	kustomizeYamlPath := filepath.Join(fdpArgoCDDir, "kustomization.yaml")
	if err := os.WriteFile(kustomizeYamlPath, []byte(kustomizationArgoYamlContent), 0644); err != nil {
		return fmt.Errorf("failed to create app.yaml: %w", err)
	}
	fmt.Printf("Created: %s\n", appYamlPath)

	// Update kustomization.yaml for ArgoCD
	kustomizePath := filepath.Join(pilotServicesArgoCDDir, "kustomization.yaml")

	//update instructions for ArgoCD
	var kustomize KustomizeConfig

	// Read existing kustomization.yaml if it exists
	if _, err := os.Stat(kustomizePath); err == nil {
		kustomizeData, err := os.ReadFile(kustomizePath)
		if err != nil {
			return fmt.Errorf("failed to read kustomization.yaml: %w", err)
		}
		if err := yaml.Unmarshal(kustomizeData, &kustomize); err != nil {
			return fmt.Errorf("failed to parse kustomization.yaml: %w", err)
		}
	} else {
		// Create new kustomization.yaml structure
		kustomize = KustomizeConfig{
			APIVersion: "kustomize.config.k8s.io/v1beta1",
			Kind:       "Kustomization",
			Resources:  []string{"project.yaml"},
		}
	}

	// Check if resource already exists
	resourceExists := false
	for _, resource := range kustomize.Resources {
		if resource == fdpName {
			resourceExists = true
			break
		}
	}

	// Copy the fdp file passed as argument to the fdps folder
	fdpsDir := filepath.Join(pilotServicesArgoCDDir, "fdps")
	target := filepath.Join(fdpsDir, fdpName+".yaml")

	moveFile(yamlDescriptorFile, target)

	fmt.Printf("FDP descriptor copied in the fdps folder\n")

	// Add new resource if it doesn't exist
	if !resourceExists {
		kustomize.Resources = append(kustomize.Resources, fdpName)
		fmt.Printf("Added %s to kustomization.yaml resources\n", fdpName)
	} else {
		fmt.Printf("Resource %s already exists in kustomization.yaml\n", fdpName)
	}

	// Write updated kustomization.yaml for ArgoCD
	kustomizeData, err := yaml.Marshal(&kustomize)
	if err != nil {
		return fmt.Errorf("failed to marshal kustomization.yaml: %w", err)
	}

	if err := os.WriteFile(kustomizePath, kustomizeData, 0644); err != nil {
		return fmt.Errorf("failed to write kustomization.yaml: %w", err)
	}
	fmt.Printf("Updated: %s\n", kustomizePath)

	// 2. Microk8s instructions

	// Create pilot-services directory if it doesn't exist
	if err := os.MkdirAll(pilotServicesK8SDir, 0755); err != nil {
		return fmt.Errorf("failed to create pilot-services directory: %w", err)
	}

	// Create FDP-specific directory
	if err := os.MkdirAll(fdpK8SDir, 0755); err != nil {
		return fmt.Errorf("failed to create FDP directory: %w", err)
	}

	kustomizationK8SYamlContent := fmt.Sprintf(`apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- base.yaml
`)

	// add kustomization.yaml for MicroK8S inside the deployment/pilot-service/fdp-descriptor.name folder
	kustomizeYamlPath = filepath.Join(fdpK8SDir, "kustomization.yaml")
	if err := os.WriteFile(kustomizeYamlPath, []byte(kustomizationK8SYamlContent), 0644); err != nil {
		return fmt.Errorf("failed to create app.yaml: %w", err)
	}
	fmt.Printf("Created: %s\n", kustomizeYamlPath)

	// Copy the deployment file passed as argument to the fdps folder
	target = filepath.Join(fdpK8SDir, "base.yaml")

	moveFile(deploymentFile, target)

	for _, resourceFile := range config.FdpDescriptor.Resources {
		source := filepath.Join(yamlDescriptorDir, resourceFile)
		target = filepath.Join(fdpK8SDir, resourceFile)
		moveFile(source, target)
	}

	fmt.Printf("FDP descriptor copied in the fdps folder\n")

	// Update kustomization.yaml for ArgoCD

	fmt.Println("Deployment successful!")
	return nil
}
