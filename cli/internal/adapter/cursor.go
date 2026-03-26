package adapter

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/digdir/skills/cli/internal/skill"
)

type cursorAdapter struct{}

func (a *cursorAdapter) Framework() Framework { return Cursor }

func (a *cursorAdapter) GlobalPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cursor", "rules")
}

func (a *cursorAdapter) ProjectPath(repoRoot string) string {
	return filepath.Join(repoRoot, ".cursor", "rules")
}

func (a *cursorAdapter) Install(s skill.Skill, targetDir string) error {
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("creating directory %s: %w", targetDir, err)
	}

	// Convert SKILL.md to .mdc format
	mdcPath := filepath.Join(targetDir, s.Name+".mdc")
	content, err := convertToMDC(s)
	if err != nil {
		return fmt.Errorf("converting to .mdc: %w", err)
	}

	if err := os.WriteFile(mdcPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", mdcPath, err)
	}

	// Also copy scripts if they exist
	if s.HasScripts {
		scriptsDir := filepath.Join(s.Path, "scripts")
		destScripts := filepath.Join(targetDir, s.Name+"-scripts")
		if err := copyDirRecursive(scriptsDir, destScripts); err != nil {
			return fmt.Errorf("copying scripts: %w", err)
		}
	}

	return nil
}

func convertToMDC(s skill.Skill) (string, error) {
	skillFile := filepath.Join(s.Path, "SKILL.md")
	f, err := os.Open(skillFile)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var b strings.Builder
	scanner := bufio.NewScanner(f)

	// Write .mdc frontmatter
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("description: %s\n", s.Description))
	b.WriteString("globs: \n")
	b.WriteString("alwaysApply: false\n")
	b.WriteString("---\n")

	// Skip original frontmatter, copy body
	inFrontmatter := false
	frontmatterDone := false
	for scanner.Scan() {
		line := scanner.Text()
		if !frontmatterDone {
			if strings.TrimSpace(line) == "---" {
				if !inFrontmatter {
					inFrontmatter = true
					continue
				}
				frontmatterDone = true
				continue
			}
			if inFrontmatter {
				continue
			}
		}
		b.WriteString(line + "\n")
	}

	return b.String(), nil
}
