package adapter

import (
	"os"
	"path/filepath"

	"github.com/digdir/skills/cli/internal/skill"
)

type codexAdapter struct{}

func (a *codexAdapter) Framework() Framework { return Codex }

func (a *codexAdapter) GlobalPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".codex", "skills")
}

func (a *codexAdapter) ProjectPath(repoRoot string) string {
	return filepath.Join(repoRoot, ".codex", "skills")
}

func (a *codexAdapter) Install(s skill.Skill, targetDir string) error {
	destDir := filepath.Join(targetDir, s.Name)
	return copySkillDir(s.Path, destDir)
}
