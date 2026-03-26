package adapter

import (
	"github.com/digdir/skills/cli/internal/skill"
)

// Framework identifies a supported agent framework.
type Framework string

const (
	ClaudeCode Framework = "claude-code"
	Cursor     Framework = "cursor"
	Copilot    Framework = "copilot"
	Codex      Framework = "codex"
)

// AllFrameworks returns all supported frameworks in display order.
func AllFrameworks() []Framework {
	return []Framework{ClaudeCode, Cursor, Copilot, Codex}
}

// DisplayName returns a human-readable name for the framework.
func (f Framework) DisplayName() string {
	switch f {
	case ClaudeCode:
		return "Claude Code"
	case Cursor:
		return "Cursor"
	case Copilot:
		return "GitHub Copilot"
	case Codex:
		return "Codex"
	default:
		return string(f)
	}
}

// Adapter installs skills for a specific agent framework.
type Adapter interface {
	// Framework returns the framework identifier.
	Framework() Framework

	// GlobalPath returns the global install path, or "" if global install is not supported.
	GlobalPath() string

	// ProjectPath returns the project-level install path for the given repo root.
	ProjectPath(repoRoot string) string

	// Install copies the skill to the given target directory.
	Install(s skill.Skill, targetDir string) error
}

// Get returns the adapter for the given framework.
func Get(f Framework) Adapter {
	switch f {
	case ClaudeCode:
		return &claudeCodeAdapter{}
	case Cursor:
		return &cursorAdapter{}
	case Copilot:
		return &copilotAdapter{}
	case Codex:
		return &codexAdapter{}
	default:
		return nil
	}
}
