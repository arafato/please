package environment

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/arafat/please/schema"
)

type Bundle struct {
	bDefs *schema.BundleDefinitions
}

func LoadBundleDefinitions(s *Environment) (*Bundle, error) {
	data, err := os.ReadFile(s.EnvironmentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var bDefs schema.BundleDefinitions
	if err := json.Unmarshal(data, &bDefs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bundle definitions: %w", err)
	}

	return &Bundle{bDefs: &bDefs}, nil
}

func (b *Bundle) SaveBundle(e *Environment) error {
	data, err := json.MarshalIndent(b.bDefs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal bundle definitions: %w", err)
	}

	err = os.WriteFile(e.EnvironmentPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (b *Bundle) IsPackageInstalled(bundleName, packageName, version string) bool {
	env, ok := b.bDefs.Bundles[bundleName]
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

func (b *Bundle) AddPackage(bundleName, packageName, version string) error {
	if bundleName == "" {
		bundleName = "default"
	}
	env, ok := b.bDefs.Bundles[bundleName]
	if !ok {
		return fmt.Errorf("bundle %q does not exist", bundleName)
	}

	if env.Packages == nil {
		env.Packages = make(map[string]string)
	}

	env.Packages[packageName] = version
	return nil
}

func (b *Bundle) DeletePackage(packageName string) error {
	bundleName := b.bDefs.ActiveBundle
	env, ok := b.bDefs.Bundles[bundleName]

	if !ok {
		return fmt.Errorf("bundle %q does not exist", bundleName)
	}

	if env.Packages == nil {
		return fmt.Errorf("bundle %q has no packages", bundleName)
	}

	delete(env.Packages, packageName)
	return nil
}

func (b *Bundle) SetActiveBundle(bundleName string) error {
	for name, _ := range b.bDefs.Bundles {
		if name == bundleName {
			b.bDefs.ActiveBundle = bundleName
			return nil
		}
	}

	return fmt.Errorf("bundle %q does not exist", bundleName)
}

func (b *Bundle) GetActiveBundle() string {
	return b.bDefs.ActiveBundle
}

func (b *Bundle) ListBundles() []string {
	names := make([]string, 0, len(b.bDefs.Bundles))
	for name, _ := range b.bDefs.Bundles {
		names = append(names, name)
	}

	return names
}

func (b *Bundle) AddBundle(bundleName, description string) error {
	for name, _ := range b.bDefs.Bundles {
		if name == bundleName {
			return fmt.Errorf("bundle %q already exists", bundleName)
		}
	}

	b.bDefs.Bundles[bundleName] = &schema.Bundle{
		Packages:    make(map[string]string),
		Description: description,
	}

	return nil
}

func (b *Bundle) GetPackageVersion(pkg string) (string, error) {
	activeBundle := b.bDefs.ActiveBundle
	bundle, _ := b.bDefs.Bundles[activeBundle]
	version, ok := bundle.Packages[pkg]
	if !ok {
		return "", fmt.Errorf("package %q does not exist in bundle %q", pkg, activeBundle)
	}

	return version, nil
}
