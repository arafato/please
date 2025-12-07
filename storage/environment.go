package storage

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/arafat/please/schema"
)

type Environment struct {
	envs *schema.EnvironmentDefinitions
}

func LoadEnvironmentDefinitions(s *Storage) (*Environment, error) {
	data, err := os.ReadFile(s.EnvironmentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var envDefs schema.EnvironmentDefinitions
	if err := json.Unmarshal(data, &envDefs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal environment definitions: %w", err)
	}

	return &Environment{envs: &envDefs}, nil
}

func (e *Environment) SaveEnvironment(s *Storage) error {
	data, err := json.MarshalIndent(e.envs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal environment definitions: %w", err)
	}

	err = os.WriteFile(s.EnvironmentPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (e *Environment) IsPackageInstalled(envName, packageName, version string) bool {
	env, ok := e.envs.Environments[envName]
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

func (e *Environment) AddPackage(envName, packageName, version string) error {
	if envName == "" {
		envName = "default"
	}
	env, ok := e.envs.Environments[envName]
	if !ok {
		return fmt.Errorf("environment %q does not exist", envName)
	}

	if env.Packages == nil {
		env.Packages = make(map[string]string)
	}

	env.Packages[packageName] = version
	return nil
}

func (e *Environment) SetActiveEnvironment(envName string) error {
	for name, _ := range e.envs.Environments {
		if name == envName {
			e.envs.ActiveEnvironment = envName
			return nil
		}
	}

	return fmt.Errorf("environment %q does not exist", envName)
}

func (e *Environment) GetActiveEnvironment() string {
	return e.envs.ActiveEnvironment
}

func (e *Environment) ListEnvironments() []string {
	names := make([]string, 0, len(e.envs.Environments))
	for name, _ := range e.envs.Environments {
		names = append(names, name)
	}

	return names
}

func (e *Environment) AddEnvironment(envName string) error {
	for name, _ := range e.envs.Environments {
		if name == envName {
			return fmt.Errorf("environment %q already exists", envName)
		}
	}

	e.envs.Environments[envName] = &schema.Environment{
		Packages: make(map[string]string),
	}

	return nil
}
