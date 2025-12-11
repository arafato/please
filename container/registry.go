package container

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/arafat/please/schema"
)

// RegistryClient handles communication with container registries
type RegistryClient struct {
	httpClient *http.Client
}

func NewRegistryClient() *RegistryClient {
	return &RegistryClient{
		httpClient: &http.Client{},
	}
}

// ListVersions fetches all versions of an image that match the filter
func (c *RegistryClient) ListVersions(ctx context.Context, manifest *schema.PackageManifest) ([]string, error) {
	if manifest == nil || manifest.VersionDiscovery == nil {
		return nil, fmt.Errorf("version discovery configuration is required")
	}

	// Parse image reference (registry/repository:tag)
	registry, repository := parseImageReference(manifest.Image)

	// Get authentication token if needed
	token, err := c.getAuthToken(ctx, registry, repository)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth token: %w", err)
	}

	// Fetch all tags from the registry
	tags, err := c.fetchTags(ctx, registry, repository, token)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tags: %w", err)
	}

	// Filter versions based on the manifest rules
	filtered, err := filterVersions(tags, manifest.VersionDiscovery.Filter)
	if err != nil {
		return nil, fmt.Errorf("failed to filter versions: %w", err)
	}

	return filtered, nil
}

// parseImageReference splits image into registry and repository
func parseImageReference(image string) (registry, repository string) {
	parts := strings.SplitN(image, "/", 2)

	// If no slash, assume Docker Hub official library image
	if len(parts) == 1 {
		return "registry-1.docker.io", "library/" + parts[0]
	}

	// Check if first part is a registry (contains . or :)
	if strings.Contains(parts[0], ".") || strings.Contains(parts[0], ":") {
		registry = parts[0]
		repository = parts[1]

		// Normalize docker.io to registry-1.docker.io
		if registry == "docker.io" {
			registry = "registry-1.docker.io"
		}

		return registry, repository
	}

	// Default to Docker Hub with user/repo format (no registry prefix)
	return "registry-1.docker.io", image
}

// getAuthToken obtains authentication token for the registry
func (c *RegistryClient) getAuthToken(ctx context.Context, registry, repository string) (string, error) {
	// For Docker Hub
	if strings.Contains(registry, "docker.io") {
		authURL := fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull", repository)
		req, err := http.NewRequestWithContext(ctx, "GET", authURL, nil)
		if err != nil {
			return "", err
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("auth request failed: %d", resp.StatusCode)
		}

		var authResp struct {
			Token string `json:"token"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
			return "", err
		}

		return authResp.Token, nil
	}

	// For other registries, might not need auth or use different mechanism
	return "", nil
}

// fetchTags retrieves all tags for an image from the registry
func (c *RegistryClient) fetchTags(ctx context.Context, registry, repository, token string) ([]string, error) {
	url := fmt.Sprintf("https://%s/v2/%s/tags/list", registry, repository)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch tags: %d", resp.StatusCode)
	}

	var tagsResp struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
		return nil, err
	}

	return tagsResp.Tags, nil
}

// filterVersions applies the pattern and exclude rules to the version list
func filterVersions(versions []string, filter schema.VersionFilter) ([]string, error) {
	var re *regexp.Regexp
	var err error

	if filter.Pattern != "" {
		re, err = regexp.Compile(filter.Pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern: %w", err)
		}
	}

	// Create exclude map for O(1) lookups
	excludeMap := make(map[string]bool)
	for _, ex := range filter.Exclude {
		excludeMap[ex] = true
	}

	var filtered []string
	for _, v := range versions {
		// Skip if in exclude list
		if excludeMap[v] {
			continue
		}

		// If pattern is defined, only include matching versions
		if re != nil && !re.MatchString(v) {
			continue
		}

		filtered = append(filtered, v)
	}

	// Sort versions semantically (latest first)
	sort.Slice(filtered, func(i, j int) bool {
		return compareVersions(filtered[i], filtered[j]) > 0
	})

	return filtered, nil
}

// compareVersions compares two semantic version strings
// Returns: 1 if v1 > v2, -1 if v1 < v2, 0 if equal
func compareVersions(v1, v2 string) int {
	// Strip 'v' prefix if present
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	// Compare each part
	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int

		if i < len(parts1) {
			n1, _ = strconv.Atoi(parts1[i])
		}
		if i < len(parts2) {
			n2, _ = strconv.Atoi(parts2[i])
		}

		if n1 > n2 {
			return 1
		}
		if n1 < n2 {
			return -1
		}
	}

	return 0
}
