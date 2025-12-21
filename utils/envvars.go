package utils

import (
	"runtime"
	"strings"
)

// Replaces ${RUNTIME_*} dynamically with according runtime data
func MakeRuntimeReplacer(version string) func(map[string]string) {
	replacements := map[string]string{
		"${RUNTIME_OS}":   runtime.GOOS,
		"${RUNTIME_ARCH}": runtime.GOARCH,
		"${RUNTIME_ITAG}": version,
	}

	return func(envVars map[string]string) {
		for key, value := range envVars {
			for placeholder, replacement := range replacements {
				if strings.Contains(value, placeholder) {
					envVars[key] = strings.ReplaceAll(value, placeholder, replacement)
				}
			}
		}
	}
}
