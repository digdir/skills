package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	repoOwner = "digdir"
	repoName  = "skills"
	apiURL    = "https://api.github.com/repos/" + repoOwner + "/" + repoName + "/releases/latest"
)

type ghRelease struct {
	TagName string    `json:"tag_name"`
	Assets  []ghAsset `json:"assets"`
}

type ghAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Update checks for the latest release and replaces the current binary if newer.
func Update(currentVersion string) error {
	fmt.Println("Checking for updates...")

	release, err := fetchLatestRelease()
	if err != nil {
		return fmt.Errorf("checking for updates: %w", err)
	}

	latest := release.TagName
	if latest == currentVersion {
		fmt.Printf("Already up to date (%s)\n", currentVersion)
		return nil
	}

	fmt.Printf("New version available: %s → %s\n", currentVersion, latest)

	assetName := binaryName()
	var downloadURL string
	for _, a := range release.Assets {
		if a.Name == assetName {
			downloadURL = a.BrowserDownloadURL
			break
		}
	}
	if downloadURL == "" {
		return fmt.Errorf("no binary found for %s/%s in release %s", runtime.GOOS, runtime.GOARCH, latest)
	}

	fmt.Printf("Downloading %s...\n", assetName)

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("finding current executable: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("resolving executable path: %w", err)
	}

	if err := downloadAndReplace(downloadURL, execPath); err != nil {
		return fmt.Errorf("updating binary: %w", err)
	}

	fmt.Printf("Updated to %s\n", latest)
	return nil
}

func fetchLatestRelease() (*ghRelease, error) {
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("no releases found (have you created a release yet?)")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %s", resp.Status)
	}

	var release ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return &release, nil
}

func binaryName() string {
	suffix := ""
	if runtime.GOOS == "windows" {
		suffix = ".exe"
	}
	return fmt.Sprintf("digdir-cli-%s-%s%s", runtime.GOOS, runtime.GOARCH, suffix)
}

func downloadAndReplace(url, destPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned %s", resp.Status)
	}

	// Write to temp file next to the binary
	dir := filepath.Dir(destPath)
	tmp, err := os.CreateTemp(dir, "digdir-cli-update-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("downloading: %w", err)
	}
	tmp.Close()

	// Make executable
	if err := os.Chmod(tmpPath, 0o755); err != nil {
		os.Remove(tmpPath)
		return err
	}

	// On Windows, we can't replace a running binary directly.
	// Rename old binary, move new one in, then delete old.
	if runtime.GOOS == "windows" {
		oldPath := destPath + ".old"
		os.Remove(oldPath) // clean up from previous update
		if err := os.Rename(destPath, oldPath); err != nil {
			os.Remove(tmpPath)
			return fmt.Errorf("renaming old binary: %w", err)
		}
		if err := os.Rename(tmpPath, destPath); err != nil {
			// Try to restore
			os.Rename(oldPath, destPath)
			os.Remove(tmpPath)
			return fmt.Errorf("moving new binary: %w", err)
		}
		// Best-effort cleanup; may fail if still locked
		os.Remove(oldPath)
		return nil
	}

	// On Unix, atomic rename works even while running
	if err := os.Rename(tmpPath, destPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("replacing binary: %w", err)
	}

	return nil
}

// CheckVersion compares current version against latest and prints a hint if outdated.
// Intended for non-blocking background-style check (but runs synchronously).
func CheckVersion(currentVersion string) {
	if currentVersion == "dev" {
		return
	}
	release, err := fetchLatestRelease()
	if err != nil {
		return // silently ignore
	}
	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(currentVersion, "v")
	if latest != current && latest != "" {
		fmt.Fprintf(os.Stderr, "\nHint: a newer version is available (%s → %s). Run 'digdir-cli update' to upgrade.\n", currentVersion, release.TagName)
	}
}
