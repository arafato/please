package environment

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/arafat/please/schema"
)

type Bundle struct {
	envs *schema.BundleDefinitions
}

func LoadBundleDefinitions(s *Environment) (*Bundle, error) {
	data, err := os.ReadFile(s.EnvironmentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var envDefs schema.BundleDefinitions
	if err := json.Unmarshal(data, &envDefs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bundle definitions: %w", err)
	}

	return &Bundle{envs: &envDefs}, nil
}

func (e *Bundle) SaveBundle(s *Environment) error {
	data, err := json.MarshalIndent(e.envs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal bundle definitions: %w", err)
	}

	err = os.WriteFile(s.EnvironmentPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (e *Bundle) IsPackageInstalled(bundleName, packageName, version string) bool {
	env, ok := e.envs.Bundles[bundleName]
	if !ok {
		return false
	}

	if env.Packages == nil {
		return false
	}

	installedVersion, ok := env.Packages[packageName]
	if !ok {
		return false
	}

	return installedVersion == version
}

func (e *Bundle) AddPackage(bundleName, packageName, version string) error {
	if bundleName == "" {
		bundleName = "default"
	}
	env, ok := e.envs.Bundles[bundleName]
	if !ok {
		return fmt.Errorf("bundle %q does not exist", bundleName)
	}

	if env.Packages == nil {
		env.Packages = make(map[string]string)
	}

	env.Packages[packageName] = version
	return nil
}

func (e *Bundle) SetActiveBundle(bundleName string) error {
	for name, _ := range e.envs.Bundles {
		if name == bundleName {
			e.envs.ActiveBundle = bundleName
			return nil
		}
	}

	return fmt.Errorf("bundle %q does not exist", bundleName)
}

func (e *Bundle) GetActiveBundle() string {
	return e.envs.ActiveBundle
}

func (e *Bundle) ListBundles() []string {
	names := make([]string, 0, len(e.envs.Bundles))
	for name, _ := range e.envs.Bundles {
		names = append(names, name)
	}

	return names
}

func (e *Bundle) AddBundle(bundleName string) error {
	for name, _ := range e.envs.Bundles {
		if name == bundleName {
			return fmt.Errorf("bundle %q already exists", bundleName)
		}
	}

	e.envs.Bundles[bundleName] = &schema.Bundle{
		Packages: make(map[string]string),
	}

	return nil
}
