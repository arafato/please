package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/arafat/please/schema"
)

func TestLoadEnvironmentDefinitions(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup
		tmpDir := t.TempDir()
		envPath := filepath.Join(tmpDir, "environments.json")

		envDefs := schema.EnvironmentDefinitions{
			ActiveEnvironment: "dev",
			Environments: map[string]*schema.Environment{
				"dev": {
					Packages: map[string]string{"pkg1": "v1.0.0"},
				},
			},
		}
		data, _ := json.Marshal(envDefs)
		err := os.WriteFile(envPath, data, 0644)
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}

		s := &Storage{EnvironmentPath: envPath}

		// Execute
		env, err := LoadEnvironmentDefinitions(s)

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if env.envs.ActiveEnvironment != "dev" {
			t.Errorf("expected ActiveEnvironment=dev, got %s", env.envs.ActiveEnvironment)
		}
	})

	t.Run("file not found", func(t *testing.T) {
		s := &Storage{EnvironmentPath: "/nonexistent/path.json"}

		env, err := LoadEnvironmentDefinitions(s)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if env != nil {
			t.Error("expected nil environment")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		tmpDir := t.TempDir()
		envPath := filepath.Join(tmpDir, "invalid.json")

		err := os.WriteFile(envPath, []byte("{invalid json}"), 0644)
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}

		s := &Storage{EnvironmentPath: envPath}

		env, err := LoadEnvironmentDefinitions(s)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if env != nil {
			t.Error("expected nil environment")
		}
	})
}

func TestSaveEnvironment(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tmpDir := t.TempDir()
		envPath := filepath.Join(tmpDir, "environments.json")

		env := &Environment{
			envs: &schema.EnvironmentDefinitions{
				ActiveEnvironment: "prod",
				Environments: map[string]*schema.Environment{
					"prod": {Packages: map[string]string{"pkg1": "v2.0.0"}},
				},
			},
		}
		s := &Storage{EnvironmentPath: envPath}

		err := env.SaveEnvironment(s)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify file was written correctly
		data, err := os.ReadFile(envPath)
		if err != nil {
			t.Fatalf("failed to read saved file: %v", err)
		}

		var loaded schema.EnvironmentDefinitions
		if err := json.Unmarshal(data, &loaded); err != nil {
			t.Fatalf("failed to unmarshal saved data: %v", err)
		}

		if loaded.ActiveEnvironment != "prod" {
			t.Errorf("expected ActiveEnvironment=prod, got %s", loaded.ActiveEnvironment)
		}
	})

	t.Run("invalid path", func(t *testing.T) {
		env := &Environment{
			envs: &schema.EnvironmentDefinitions{},
		}
		s := &Storage{EnvironmentPath: "/invalid/path/file.json"}

		err := env.SaveEnvironment(s)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestAddPackage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := &Environment{
			envs: &schema.EnvironmentDefinitions{
				Environments: map[string]*schema.Environment{
					"dev": {Packages: map[string]string{}},
				},
			},
		}

		err := env.AddPackage("dev", "newpkg", "v1.0.0")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if env.envs.Environments["dev"].Packages["newpkg"] != "v1.0.0" {
			t.Error("package was not added correctly")
		}
	})

	t.Run("environment not found", func(t *testing.T) {
		env := &Environment{
			envs: &schema.EnvironmentDefinitions{
				Environments: map[string]*schema.Environment{},
			},
		}

		err := env.AddPackage("nonexistent", "pkg", "v1.0.0")

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("nil packages map", func(t *testing.T) {
		env := &Environment{
			envs: &schema.EnvironmentDefinitions{
				Environments: map[string]*schema.Environment{
					"dev": {Packages: nil},
				},
			},
		}

		err := env.AddPackage("dev", "pkg", "v1.0.0")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if env.envs.Environments["dev"].Packages["pkg"] != "v1.0.0" {
			t.Error("package was not added correctly")
		}
	})
}

func TestSetActiveEnvironment(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := &Environment{
			envs: &schema.EnvironmentDefinitions{
				ActiveEnvironment: "dev",
				Environments: map[string]*schema.Environment{
					"dev":  {},
					"prod": {},
				},
			},
		}

		err := env.SetActiveEnvironment("prod")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if env.envs.ActiveEnvironment != "prod" {
			t.Errorf("expected ActiveEnvironment=prod, got %s", env.envs.ActiveEnvironment)
		}
	})

	t.Run("environment not found", func(t *testing.T) {
		env := &Environment{
			envs: &schema.EnvironmentDefinitions{
				Environments: map[string]*schema.Environment{
					"dev": {},
				},
			},
		}

		err := env.SetActiveEnvironment("nonexistent")

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestGetActiveEnvironment(t *testing.T) {
	env := &Environment{
		envs: &schema.EnvironmentDefinitions{
			ActiveEnvironment: "staging",
		},
	}

	active := env.GetActiveEnvironment()

	if active != "staging" {
		t.Errorf("expected staging, got %s", active)
	}
}

func TestListEnvironments(t *testing.T) {
	env := &Environment{
		envs: &schema.EnvironmentDefinitions{
			Environments: map[string]*schema.Environment{
				"dev":     {},
				"staging": {},
				"prod":    {},
			},
		},
	}

	names := env.ListEnvironments()

	if len(names) != 3 {
		t.Errorf("expected 3 environments, got %d", len(names))
	}

	// Check all expected names are present
	expected := map[string]bool{"dev": false, "staging": false, "prod": false}
	for _, name := range names {
		if _, ok := expected[name]; ok {
			expected[name] = true
		}
	}
	for name, found := range expected {
		if !found {
			t.Errorf("expected environment %s not found", name)
		}
	}
}

func TestAddEnvironment(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := &Environment{
			envs: &schema.EnvironmentDefinitions{
				Environments: map[string]*schema.Environment{
					"dev": {},
				},
			},
		}

		err := env.AddEnvironment("prod")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if _, ok := env.envs.Environments["prod"]; !ok {
			t.Error("environment was not added")
		}
		if env.envs.Environments["prod"].Packages == nil {
			t.Error("packages map should be initialized")
		}
	})

	t.Run("environment already exists", func(t *testing.T) {
		env := &Environment{
			envs: &schema.EnvironmentDefinitions{
				Environments: map[string]*schema.Environment{
					"dev": {},
				},
			},
		}

		err := env.AddEnvironment("dev")

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
