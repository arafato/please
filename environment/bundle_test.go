package environment

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

		envDefs := schema.BundleDefinitions{
			ActiveBundle: "dev",
			Bundles: map[string]*schema.Bundle{
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

		s := &Environment{EnvironmentPath: envPath}

		// Execute
		env, err := LoadBundleDefinitions(s)

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if env.bDefs.ActiveBundle != "dev" {
			t.Errorf("expected ActiveEnvironment=dev, got %s", env.bDefs.ActiveBundle)
		}
	})

	t.Run("file not found", func(t *testing.T) {
		s := &Environment{EnvironmentPath: "/nonexistent/path.json"}

		env, err := LoadBundleDefinitions(s)

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

		s := &Environment{EnvironmentPath: envPath}

		env, err := LoadBundleDefinitions(s)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if env != nil {
			t.Error("expected nil environment")
		}
	})
}

func TestSaveBundle(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tmpDir := t.TempDir()
		envPath := filepath.Join(tmpDir, "environments.json")

		env := &Bundle{
			bDefs: &schema.BundleDefinitions{
				ActiveBundle: "prod",
				Bundles: map[string]*schema.Bundle{
					"prod": {Packages: map[string]string{"pkg1": "v2.0.0"}},
				},
			},
		}
		s := &Environment{EnvironmentPath: envPath}

		err := env.SaveBundle(s)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify file was written correctly
		data, err := os.ReadFile(envPath)
		if err != nil {
			t.Fatalf("failed to read saved file: %v", err)
		}

		var loaded schema.BundleDefinitions
		if err := json.Unmarshal(data, &loaded); err != nil {
			t.Fatalf("failed to unmarshal saved data: %v", err)
		}

		if loaded.ActiveBundle != "prod" {
			t.Errorf("expected ActiveEnvironment=prod, got %s", loaded.ActiveBundle)
		}
	})

	t.Run("invalid path", func(t *testing.T) {
		env := &Bundle{
			bDefs: &schema.BundleDefinitions{},
		}
		s := &Environment{EnvironmentPath: "/invalid/path/file.json"}

		err := env.SaveBundle(s)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestAddPackage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := &Bundle{
			bDefs: &schema.BundleDefinitions{
				Bundles: map[string]*schema.Bundle{
					"dev": {Packages: map[string]string{}},
				},
			},
		}

		err := env.AddPackage("dev", "newpkg", "v1.0.0")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if env.bDefs.Bundles["dev"].Packages["newpkg"] != "v1.0.0" {
			t.Error("package was not added correctly")
		}
	})

	t.Run("bundle not found", func(t *testing.T) {
		env := &Bundle{
			bDefs: &schema.BundleDefinitions{
				Bundles: map[string]*schema.Bundle{},
			},
		}

		err := env.AddPackage("nonexistent", "pkg", "v1.0.0")

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("nil packages map", func(t *testing.T) {
		env := &Bundle{
			bDefs: &schema.BundleDefinitions{
				Bundles: map[string]*schema.Bundle{
					"dev": {Packages: nil},
				},
			},
		}

		err := env.AddPackage("dev", "pkg", "v1.0.0")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if env.bDefs.Bundles["dev"].Packages["pkg"] != "v1.0.0" {
			t.Error("package was not added correctly")
		}
	})
}

func TestSetActiveBundle(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := &Bundle{
			bDefs: &schema.BundleDefinitions{
				ActiveBundle: "dev",
				Bundles: map[string]*schema.Bundle{
					"dev":  {},
					"prod": {},
				},
			},
		}

		err := env.SetActiveBundle("prod")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if env.bDefs.ActiveBundle != "prod" {
			t.Errorf("expected ActiveBundle=prod, got %s", env.bDefs.ActiveBundle)
		}
	})

	t.Run("bundle not found", func(t *testing.T) {
		env := &Bundle{
			bDefs: &schema.BundleDefinitions{
				Bundles: map[string]*schema.Bundle{
					"dev": {},
				},
			},
		}

		err := env.SetActiveBundle("nonexistent")

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestGetActiveBundle(t *testing.T) {
	env := &Bundle{
		bDefs: &schema.BundleDefinitions{
			ActiveBundle: "staging",
		},
	}

	active := env.GetActiveBundle()

	if active != "staging" {
		t.Errorf("expected staging, got %s", active)
	}
}

func TestListBundle(t *testing.T) {
	env := &Bundle{
		bDefs: &schema.BundleDefinitions{
			Bundles: map[string]*schema.Bundle{
				"dev":     {},
				"staging": {},
				"prod":    {},
			},
		},
	}

	names := env.ListBundles()

	if len(names) != 3 {
		t.Errorf("expected 3 bundles, got %d", len(names))
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
			t.Errorf("expected bundle %s not found", name)
		}
	}
}

func TestAddBundle(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := &Bundle{
			bDefs: &schema.BundleDefinitions{
				Bundles: map[string]*schema.Bundle{
					"dev": {},
				},
			},
		}

		err := env.AddBundle("prod", "Production environment")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if _, ok := env.bDefs.Bundles["prod"]; !ok {
			t.Error("bundle was not added")
		}
		if env.bDefs.Bundles["prod"].Packages == nil {
			t.Error("packages map should be initialized")
		}
	})

	t.Run("bundle already exists", func(t *testing.T) {
		env := &Bundle{
			bDefs: &schema.BundleDefinitions{
				Bundles: map[string]*schema.Bundle{
					"dev": {},
				},
			},
		}

		err := env.AddBundle("dev", "Development environment")

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
