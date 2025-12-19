/*
 * Core Initialization
 */
package environment

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/arafat/please/artifacts"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

const (
	defaultSourceURL = "https://please.oarafat.workers.dev/manifest-core.tar.gz"
	sourcesFile      = "sources"
	pleaseDir        = ".please"
)

type Environment struct {
	homePath         string
	PleasePath       string
	manifestPath     string
	ManifestCoreFile string
	EnvironmentPath  string
	BinPath          string
	VersionsPath     string
	Platform         string
	Arch             string
	OS               string
}

func New() *Environment {
	homeDir, _ := os.UserHomeDir()
	arch := strings.ToLower(runtime.GOARCH)
	os := "linux" // no darwin images available strings.ToLower(runtime.GOOS)

	return &Environment{
		homePath:         homeDir,
		PleasePath:       fmt.Sprintf("%s/%s", homeDir, pleaseDir),
		manifestPath:     fmt.Sprintf("%s/%s/%s", homeDir, pleaseDir, "manifests"),
		ManifestCoreFile: fmt.Sprintf("%s/%s/%s/%s", homeDir, pleaseDir, "manifests", "manifest-core.tar.gz"),
		EnvironmentPath:  fmt.Sprintf("%s/%s/%s", homeDir, pleaseDir, "env.json"),
		BinPath:          fmt.Sprintf("%s/%s/%s", homeDir, pleaseDir, "bin"),
		VersionsPath:     fmt.Sprintf("%s/%s/%s", homeDir, pleaseDir, "versions"),
		Platform:         fmt.Sprintf("%s/%s", os, arch),
		Arch:             arch,
		OS:               os,
	}
}

func (e *Environment) DeployArtifact(d artifacts.Deployable, pkg, executable, version string) (string, error) {
	installationFullPath := fmt.Sprintf("%s/%s/%s/%s.sh", e.VersionsPath, pkg, version, executable)
	installationPath := fmt.Sprintf("%s/%s/%s", e.VersionsPath, pkg, version)

	err := os.MkdirAll(installationPath, 0755)
	if err != nil {
		return "", fmt.Errorf("Error creating directories:%w", err)
	}
	if err := d.Deploy(installationFullPath); err != nil {
		return "", fmt.Errorf("failed to deploy script: %w", err)
	}
	return installationFullPath, nil
}

func (e *Environment) CreateSymlink(pkg, executable, targetVersion string) error {
	targetPath := fmt.Sprintf("%s/%s/%s/%s.sh", e.VersionsPath, pkg, targetVersion, executable)
	symlinkPath := fmt.Sprintf("%s/%s", e.BinPath, executable)

	// Remove existing file/symlink if it exists
	if _, err := os.Lstat(symlinkPath); err == nil {
		err := os.Remove(symlinkPath)
		if err != nil {
			return fmt.Errorf("Error removing existing symlink:%w", err)
		}
	}

	if err := os.Symlink(targetPath, symlinkPath); err != nil {
		return fmt.Errorf("failed to create symlink in %s: %w", symlinkPath, err)
	}
	return nil
}

func (e *Environment) DeleteArtifact(pkg, version string) error {
	installationPath := fmt.Sprintf("%s/%s/%s", e.VersionsPath, pkg, version)
	if err := os.RemoveAll(installationPath); err != nil {
		return fmt.Errorf("Error removing artifact:%w", err)
	}
	return nil
}

func (e *Environment) DeleteSymlink(executable string) error {
	symlinkPath := fmt.Sprintf("%s/%s", e.BinPath, executable)
	if err := os.Remove(symlinkPath); err != nil {
		return fmt.Errorf("Error removing existing symlink:%w", err)
	}
	return nil
}

func (e *Environment) GetManifestPaths() ([]string, error) {
	manifests, err := os.ReadDir(e.manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest directory: %w", err)
	}

	if len(manifests) == 0 {
		return nil, errors.New("No manifest files found. Please run '$ please update'.")
	}

	var files []string
	for _, item := range manifests {
		if !item.IsDir() && strings.HasSuffix(item.Name(), ".tar.gz") {
			files = append(files, filepath.Join(e.manifestPath, item.Name()))
		}
	}
	return files, nil
}

func (e *Environment) LoadSources() ([]string, error) {
	file, err := os.Open(e.SourcesPath())
	if err != nil {
		return nil, fmt.Errorf("failed to open sources file: %w", err)
	}
	defer file.Close()

	var sources []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			sources = append(sources, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading sources file: %w", err)
	}

	if len(sources) == 0 {
		return nil, fmt.Errorf("no sources found in %s", e.SourcesPath())
	}

	return sources, nil
}

func (e *Environment) DownloadManifestFiles(urls []string) {

	p := mpb.New(mpb.WithWidth(60))
	var wg sync.WaitGroup
	for _, url := range urls {
		// TODO: Check ETag - if no new version skip download
		wg.Add(1)
		fileName := path.Base(url)
		go downloadManifest(url, e.ManifestPath(fileName), p, &wg)
	}
	wg.Wait()
	p.Wait()
}

func (e *Environment) SourcesPath() string {
	return filepath.Join(e.PleasePath, sourcesFile)
}

func (e *Environment) ManifestPath(manifestName string) string {
	return filepath.Join(e.manifestPath, manifestName)
}

func (e *Environment) IsInitialized() bool {
	if stat, err := os.Stat(e.PleasePath); err == nil && stat.IsDir() {
		return true
	}
	return false
}

func (e *Environment) Initialize() error {
	if err := os.MkdirAll(e.PleasePath, 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(e.manifestPath, 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(e.BinPath, 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(e.VersionsPath, 0755); err != nil {
		return err
	}

	if err := e.createSources(); err != nil {
		return err
	}

	return nil
}

func (e *Environment) createSources() error {
	path := e.SourcesPath()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		content := fmt.Sprintf("# please package sources\n# Add one URL per line\n\n%s\n", defaultSourceURL)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create sources file: %w", err)
		}
	}

	return nil
}

func downloadManifest(url, filename string, p *mpb.Progress, wg *sync.WaitGroup) {
	defer wg.Done()

	// Create output file
	out, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer out.Close()

	// Perform request
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error downloading:", err)
		return
	}
	defer resp.Body.Close()

	// Get content length (if provided)
	size := resp.ContentLength
	if size <= 0 {
		size = 0 // unknown size
	}

	// Create progress bar
	bar := p.AddBar(
		size,
		mpb.PrependDecorators(
			decor.Name(filepath.Base(filename)+" "),
			decor.CountersKibiByte("% .2f / % .2f"),
		),
		mpb.AppendDecorators(
			decor.Percentage(),
		),
	)

	// Wrap reader with proxy for bar
	proxyReader := bar.ProxyReader(resp.Body)
	defer proxyReader.Close()

	// Copy data to file
	_, err = io.Copy(out, proxyReader)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}
