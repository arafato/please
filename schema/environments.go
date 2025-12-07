package schema

type EnvironmentDefinitions struct {
	Environments      map[string]*Environment `json:"environments"`
	ActiveEnvironment string                  `json:"activeEnvironment"`
}

type Environment struct {
	Description string            `json:"description"`
	Packages    map[string]string `json:"packages"`
}

func NewDefaultEnvironment() *EnvironmentDefinitions {
	return &EnvironmentDefinitions{
		ActiveEnvironment: "default",
		Environments: map[string]*Environment{
			"default": {
				Description: "Default environment created by please",
				Packages:    map[string]string{},
			},
		},
	}
}
