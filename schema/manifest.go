package schema

// --- Main Manifest Structure ---
type PackageManifest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Homepage    string   `json:"homepage"`
	License     string   `json:"license"`
	Categories  []string `json:"categories"`
	Image       string   `json:"image"`

	// If both are present in JSON, both will be populated.
	// Application logic will enforce precedence (Versions > VersionDiscovery).
	Versions         []string          `json:"versions,omitempty"`
	VersionDiscovery *VersionDiscovery `json:"version_discovery,omitempty"`

	DefaultVersion  string        `json:"default_version"`
	Script          string        `json:"script"`
	Platforms       []string      `json:"platforms"`
	ApplicationArgs []string      `json:"application_args"`
	ContainerArgs   ContainerArgs `json:"container_args"`
}

// ContainerArgs maps directly to the container_args JSON object.
type ContainerArgs struct {
	DNS             []string `json:"dns"`
	WorkDir         string   `json:"workdir"`
	Volumes         []string `json:"volumes"`
	AdditionalFlags []string `json:"additional_flags"`
}

// VersionFilter defines the pattern and exclude rules for version discovery.
type VersionFilter struct {
	Pattern string   `json:"pattern"` // e.g. "^[0-9]+\\.[0-9]+\\.[0-9]+$"
	Exclude []string `json:"exclude"` // e.g. ["latest", "edge", "rc"]
}

// VersionDiscovery encapsulates the version filtering rules.
type VersionDiscovery struct {
	Filter VersionFilter `json:"filter"`
}
