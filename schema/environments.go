package schema

type BundleDefinitions struct {
	Bundles      map[string]*Bundle `json:"environments"`
	ActiveBundle string             `json:"activeEnvironment"`
}

type Bundle struct {
	Description string            `json:"description"`
	Packages    map[string]string `json:"packages"`
}

func NewDefaultBundle() *BundleDefinitions {
	return &BundleDefinitions{
		ActiveBundle: "default",
		Bundles: map[string]*Bundle{
			"default": {
				Description: "Default bundle created by please",
				Packages:    map[string]string{},
			},
		},
	}
}
