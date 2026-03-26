package skill

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

const (
	defaultRepo   = "https://github.com/digdir/skills.git"
	cacheTTL      = 1 * time.Hour
	skillsSubdir  = "skills"
)

// Source resolves the path to the skills directory, either from a local path or by cloning from GitHub.
type Source struct {
	LocalPath string // if set, use this local path instead of cloning
	RepoURL   string // GitHub repo URL (defaults to digdir/skills)
}

// Resolve returns the absolute path to the "skills/" directory containing skill definitions.
func (s *Source) Resolve() (string, error) {
	if s.LocalPath != "" {
		return s.resolveLocal()
	}
	return s.resolveGitHub()
}

func (s *Source) resolveLocal() (string, error) {
	skillsDir := filepath.Join(s.LocalPath, skillsSubdir)
	info, err := os.Stat(skillsDir)
	if err != nil {
		return "", fmt.Errorf("local skills directory not found: %s", skillsDir)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("not a directory: %s", skillsDir)
	}
	return skillsDir, nil
}

func (s *Source) resolveGitHub() (string, error) {
	cacheDir := s.cacheDir()
	repoDir := filepath.Join(cacheDir, "skills-repo")
	skillsDir := filepath.Join(repoDir, skillsSubdir)

	// Check if cache is fresh
	if info, err := os.Stat(filepath.Join(repoDir, ".git")); err == nil {
		if time.Since(info.ModTime()) < cacheTTL {
			return skillsDir, nil
		}
		// Cache is stale, try to pull
		if err := gitPull(repoDir); err == nil {
			return skillsDir, nil
		}
		// Pull failed, remove and re-clone
		os.RemoveAll(repoDir)
	}

	repoURL := s.RepoURL
	if repoURL == "" {
		repoURL = defaultRepo
	}

	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return "", fmt.Errorf("creating cache directory: %w", err)
	}

	if err := gitClone(repoURL, repoDir); err != nil {
		return "", fmt.Errorf("cloning repository: %w", err)
	}

	return skillsDir, nil
}

func (s *Source) cacheDir() string {
	if runtime.GOOS == "windows" {
		if appData := os.Getenv("LOCALAPPDATA"); appData != "" {
			return filepath.Join(appData, "digdir-cli", "cache")
		}
	}
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		return filepath.Join(xdg, "digdir-cli")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "digdir-cli")
}

func gitClone(url, dest string) error {
	cmd := exec.Command("git", "clone", "--depth", "1", url, dest)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func gitPull(repoDir string) error {
	cmd := exec.Command("git", "-C", repoDir, "pull", "--ff-only")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
