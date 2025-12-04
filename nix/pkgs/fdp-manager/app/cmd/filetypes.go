package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type FdpConfig struct {
	FdpDescriptor struct {
		Name        string   `yaml:"name"`
		Version     string   `yaml:"version"`
		Description string   `yaml:"description"`
		Deployment  string   `yaml:"deployment"`
		Resources   []string `yaml:"resources"`
	} `yaml:"fdp-descriptor"`
}

type KustomizeConfig struct {
	APIVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Resources  []string `yaml:"resources"`
}

// AppConfig represents the structure of config.yaml
type AppConfig struct {
	Paths struct {
		Node     string `yaml:"node"`
		ArgoCD   string `yaml:"argocd"`
		MicroK8S string `yaml:"microk8s"`
	} `yaml:"paths"`
}

var Config AppConfig

// LoadConfig reads config.yaml on application startup
func LoadConfig() error {
	data, err := os.ReadFile("config/config.yaml")
	if err != nil {
		return fmt.Errorf("failed to read config.yaml: %w", err)
	}

	if err := yaml.Unmarshal(data, &Config); err != nil {
		return fmt.Errorf("failed to parse config.yaml: %w", err)
	}

	return nil
}

func removeFromKustomizationResources(baseDir, fdpName string) error {
	kustPath := filepath.Join(baseDir, "kustomization.yaml")

	data, err := os.ReadFile(kustPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// If no kustomization file exists, just log and continue
			fmt.Printf("kustomization.yaml not found at %s, skipping resources cleanup\n", kustPath)
			return nil
		}
		return fmt.Errorf("reading kustomization.yaml: %w", err)
	}

	var k KustomizeConfig
	if err := yaml.Unmarshal(data, &k); err != nil {
		return fmt.Errorf("parsing kustomization.yaml: %w", err)
	}

	if len(k.Resources) == 0 {
		// nothing to remove
		fmt.Println("kustomization.yaml has no resources, nothing to clean up")
		return nil
	}

	// Filter out entries equal to fdpName
	newResources := make([]string, 0, len(k.Resources))
	removed := false
	for _, r := range k.Resources {
		if r == fdpName {
			removed = true
			continue
		}
		newResources = append(newResources, r)
	}
	k.Resources = newResources

	if !removed {
		fmt.Printf("Resource %q not found in kustomization.yaml resources, nothing to remove\n", fdpName)
		return nil
	}

	out, err := yaml.Marshal(&k)
	if err != nil {
		return fmt.Errorf("marshalling updated kustomization.yaml: %w", err)
	}

	if err := os.WriteFile(kustPath, out, 0o644); err != nil {
		return fmt.Errorf("writing updated kustomization.yaml: %w", err)
	}

	fmt.Printf("Removed resource %q from kustomization.yaml\n", fdpName)
	return nil
}

// helper to create an io.Reader from a byte slice
func fileBytesReader(b []byte) io.Reader {
	return &byteReader{b: b}
}

type byteReader struct {
	b []byte
}

func (r *byteReader) Read(p []byte) (int, error) {
	if len(r.b) == 0 {
		return 0, io.EOF
	}
	n := copy(p, r.b)
	r.b = r.b[n:]
	return n, nil
}

func moveFile(source, destination string) error {

	inputBytes, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("reading yaml file: %w", err)
	}

	// Create (or truncate) the target file
	outFile, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("creating target file %s: %w", destination, err)
	}
	defer outFile.Close()

	// Copy the original YAML content into the new file
	if _, err := io.Copy(outFile, fileBytesReader(inputBytes)); err != nil {
		return fmt.Errorf("writing to target file %s: %w", destination, err)
	}

	return nil
}
