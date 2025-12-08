/*
 * Core Initialization
 */
package storage

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
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

type Storage struct {
	homePath         string
	PleasePath       string
	manifestPath     string
	ManifestCoreFile string
	EnvironmentPath  string
	BinPath          string
	VersionsPath     string
}

func New() *Storage {
	homeDir, _ := os.UserHomeDir()
	return &Storage{
		homePath:         homeDir,
		PleasePath:       fmt.Sprintf("%s/%s", homeDir, pleaseDir),
		manifestPath:     fmt.Sprintf("%s/%s/%s", homeDir, pleaseDir, "manifests"),
		ManifestCoreFile: fmt.Sprintf("%s/%s/%s/%s", homeDir, pleaseDir, "manifests", "manifest-core.tar.gz"),
		EnvironmentPath:  fmt.Sprintf("%s/%s/%s", homeDir, pleaseDir, "env.json"),
		BinPath:          fmt.Sprintf("%s/%s/%s", homeDir, pleaseDir, "bin"),
		VersionsPath:     fmt.Sprintf("%s/%s/%s", homeDir, pleaseDir, "versions"),
	}
}

func (s *Storage) DeployScript(d artifacts.Deployable, pkg, version string) (string, error) {
	installationFullPath := fmt.Sprintf("%s/%s/%s/%s.sh", s.VersionsPath, pkg, version, pkg)
	installationPath := fmt.Sprintf("%s/%s/%s", s.VersionsPath, pkg, version)

	err := os.MkdirAll(installationPath, 0755)
	if err != nil {
		return "", fmt.Errorf("Error creating directories:%w", err)
	}
	if err := d.WriteScript(installationFullPath); err != nil {
		return "", fmt.Errorf("failed to deploy script: %w", err)
	}
	return installationFullPath, nil
}

func (s *Storage) CreateSymlink(pkg, targetVersion string) error {
	targetPath := fmt.Sprintf("%s/%s/%s/%s.sh", s.VersionsPath, pkg, targetVersion, pkg)
	symlinkPath := fmt.Sprintf("%s/%s", s.BinPath, pkg)

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

func (s *Storage) GetManifestPaths() ([]string, error) {
	manifests, err := os.ReadDir(s.manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest directory: %w", err)
	}

	if len(manifests) == 0 {
		return nil, errors.New("No manifest files found. Please run '$ please update'.")
	}

	var files []string
	for _, item := range manifests {
		if !item.IsDir() && strings.HasSuffix(item.Name(), ".tar.gz") {
			files = append(files, filepath.Join(s.manifestPath, item.Name()))
		}
	}
	return files, nil
}

func (s *Storage) LoadSources() ([]string, error) {
	file, err := os.Open(s.SourcesPath())
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
		return nil, fmt.Errorf("no sources found in %s", s.SourcesPath())
	}

	return sources, nil
}

func (s *Storage) DownloadManifestFiles(urls []string) {

	p := mpb.New(mpb.WithWidth(60))
	var wg sync.WaitGroup
	for _, url := range urls {
		// TODO: Check ETag - if no new version skip download
		wg.Add(1)
		fileName := path.Base(url)
		go downloadManifest(url, s.ManifestPath(fileName), p, &wg)
	}
	wg.Wait()
	p.Wait()
}

func (s *Storage) SourcesPath() string {
	return filepath.Join(s.PleasePath, sourcesFile)
}

func (s *Storage) ManifestPath(manifestName string) string {
	return filepath.Join(s.manifestPath, manifestName)
}

func (s *Storage) IsInitialized() bool {
	if stat, err := os.Stat(s.PleasePath); err == nil && stat.IsDir() {
		return true
	}
	return false
}

func (s *Storage) Initialize() error {
	if err := os.MkdirAll(s.PleasePath, 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(s.manifestPath, 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(s.BinPath, 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(s.VersionsPath, 0755); err != nil {
		return err
	}

	if err := s.createSources(); err != nil {
		return err
	}

	return nil
}

func (s *Storage) createSources() error {
	path := s.SourcesPath()

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
