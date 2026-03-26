package adapter

import (
	"os"
	"path/filepath"

	"github.com/digdir/skills/cli/internal/skill"
)

type claudeCodeAdapter struct{}

func (a *claudeCodeAdapter) Framework() Framework { return ClaudeCode }

func (a *claudeCodeAdapter) GlobalPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "skills")
}

func (a *claudeCodeAdapter) ProjectPath(repoRoot string) string {
	return filepath.Join(repoRoot, ".claude", "skills")
}

func (a *claudeCodeAdapter) Install(s skill.Skill, targetDir string) error {
	destDir := filepath.Join(targetDir, s.Name)
	return copySkillDir(s.Path, destDir)
}
