package adapter

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/digdir/skills/cli/internal/skill"
)

type copilotAdapter struct{}

func (a *copilotAdapter) Framework() Framework { return Copilot }

func (a *copilotAdapter) GlobalPath() string {
	return "" // Copilot only supports project-level instructions
}

func (a *copilotAdapter) ProjectPath(repoRoot string) string {
	return filepath.Join(repoRoot, ".github")
}

func (a *copilotAdapter) Install(s skill.Skill, targetDir string) error {
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("creating directory %s: %w", targetDir, err)
	}

	instructionsFile := filepath.Join(targetDir, "copilot-instructions.md")

	// Read skill content (body only, skip frontmatter)
	body, err := readSkillBody(filepath.Join(s.Path, "SKILL.md"))
	if err != nil {
		return fmt.Errorf("reading skill body: %w", err)
	}

	sectionHeader := fmt.Sprintf("<!-- digdir-skill: %s -->", s.Name)
	sectionEnd := fmt.Sprintf("<!-- /digdir-skill: %s -->", s.Name)
	newSection := fmt.Sprintf("%s\n%s\n%s\n", sectionHeader, strings.TrimSpace(body), sectionEnd)

	// Read existing file if it exists
	existing := ""
	if data, err := os.ReadFile(instructionsFile); err == nil {
		existing = string(data)
	}

	// Replace existing section or append
	if strings.Contains(existing, sectionHeader) {
		// Replace the existing section
		start := strings.Index(existing, sectionHeader)
		end := strings.Index(existing, sectionEnd)
		if end != -1 {
			end += len(sectionEnd)
			// Include trailing newline if present
			if end < len(existing) && existing[end] == '\n' {
				end++
			}
			existing = existing[:start] + newSection + existing[end:]
		}
	} else {
		// Append
		if existing != "" && !strings.HasSuffix(existing, "\n") {
			existing += "\n"
		}
		if existing != "" {
			existing += "\n"
		}
		existing += newSection
	}

	if err := os.WriteFile(instructionsFile, []byte(existing), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", instructionsFile, err)
	}

	// Copy scripts if they exist
	if s.HasScripts {
		scriptsDir := filepath.Join(s.Path, "scripts")
		destScripts := filepath.Join(targetDir, "copilot-skills", s.Name, "scripts")
		if err := copyDirRecursive(scriptsDir, destScripts); err != nil {
			return fmt.Errorf("copying scripts: %w", err)
		}
	}

	return nil
}

func readSkillBody(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var b strings.Builder
	scanner := bufio.NewScanner(f)

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
