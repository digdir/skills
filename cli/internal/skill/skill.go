package skill

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Skill represents a discovered skill from the skills repository.
type Skill struct {
	Name             string
	Description      string
	ShortDescription string
	Path             string // absolute path to the skill directory
	HasScripts       bool
}

// Discover finds all skills in the given skills directory by looking for SKILL.md files.
func Discover(skillsDir string) ([]Skill, error) {
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil, fmt.Errorf("reading skills directory: %w", err)
	}

	var skills []Skill
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillDir := filepath.Join(skillsDir, entry.Name())
		skillFile := filepath.Join(skillDir, "SKILL.md")
		if _, err := os.Stat(skillFile); err != nil {
			continue
		}

		s, err := parseSkillFile(skillFile, skillDir)
		if err != nil {
			continue // skip unparseable skills
		}
		skills = append(skills, s)
	}
	return skills, nil
}

func parseSkillFile(path, skillDir string) (Skill, error) {
	f, err := os.Open(path)
	if err != nil {
		return Skill{}, err
	}
	defer f.Close()

	s := Skill{Path: skillDir}

	scanner := bufio.NewScanner(f)
	inFrontmatter := false
	inMetadata := false

	for scanner.Scan() {
		line := scanner.Text()

		if !inFrontmatter && strings.TrimSpace(line) == "---" {
			inFrontmatter = true
			continue
		}
		if inFrontmatter && strings.TrimSpace(line) == "---" {
			break // end of frontmatter
		}
		if !inFrontmatter {
			continue
		}

		// Simple YAML parsing for the fields we need
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "name:") {
			s.Name = unquote(strings.TrimPrefix(trimmed, "name:"))
		} else if strings.HasPrefix(trimmed, "description:") {
			s.Description = unquote(strings.TrimPrefix(trimmed, "description:"))
		} else if strings.HasPrefix(trimmed, "metadata:") {
			inMetadata = true
		} else if inMetadata && strings.HasPrefix(trimmed, "short-description:") {
			s.ShortDescription = unquote(strings.TrimPrefix(trimmed, "short-description:"))
		} else if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			inMetadata = false
		}
	}

	if s.Name == "" {
		s.Name = filepath.Base(skillDir)
	}
	if s.ShortDescription == "" {
		s.ShortDescription = s.Description
	}

	// Check for scripts directory
	scriptsDir := filepath.Join(skillDir, "scripts")
	if info, err := os.Stat(scriptsDir); err == nil && info.IsDir() {
		s.HasScripts = true
	}

	return s, nil
}

func unquote(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	if len(s) >= 2 && s[0] == '\'' && s[len(s)-1] == '\'' {
		s = s[1 : len(s)-1]
	}
	return s
}
