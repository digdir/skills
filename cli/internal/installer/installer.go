package installer

import (
	"fmt"

	"github.com/digdir/skills/cli/internal/adapter"
	"github.com/digdir/skills/cli/internal/skill"
)

// Result represents the outcome of installing a single skill to a single target.
type Result struct {
	Skill     skill.Skill
	TargetDir string
	Err       error
}

// Config holds the installation parameters.
type Config struct {
	Framework adapter.Framework
	Skills    []skill.Skill
	Global    bool     // install to global path
	RepoPaths []string // install to these project roots (if not global)
}

// Install provisions the selected skills to the configured targets.
func Install(cfg Config) []Result {
	a := adapter.Get(cfg.Framework)
	if a == nil {
		return []Result{{Err: fmt.Errorf("unsupported framework: %s", cfg.Framework)}}
	}

	var results []Result

	if cfg.Global {
		globalPath := a.GlobalPath()
		if globalPath == "" {
			results = append(results, Result{
				Err: fmt.Errorf("%s does not support global installation", cfg.Framework.DisplayName()),
			})
		} else {
			for _, s := range cfg.Skills {
				err := a.Install(s, globalPath)
				results = append(results, Result{
					Skill:     s,
					TargetDir: globalPath,
					Err:       err,
				})
			}
		}
	}

	for _, repoPath := range cfg.RepoPaths {
		projectPath := a.ProjectPath(repoPath)
		for _, s := range cfg.Skills {
			err := a.Install(s, projectPath)
			results = append(results, Result{
				Skill:     s,
				TargetDir: projectPath,
				Err:       err,
			})
		}
	}

	return results
}
