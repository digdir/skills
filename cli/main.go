package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/digdir/skills/cli/internal/adapter"
	"github.com/digdir/skills/cli/internal/installer"
	"github.com/digdir/skills/cli/internal/skill"
	"github.com/digdir/skills/cli/internal/tui"
	"github.com/digdir/skills/cli/internal/updater"
)

// Set via -ldflags at build time.
var version = "dev"

func main() {
	args := os.Args[1:]

	// Extract global flags manually so they work with any subcommand position
	source := ""
	showVersion := false
	var filtered []string
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--version" || args[i] == "-version" || args[i] == "-v":
			showVersion = true
		case (args[i] == "--source" || args[i] == "-source") && i+1 < len(args):
			source = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--source="):
			source = strings.TrimPrefix(args[i], "--source=")
		case strings.HasPrefix(args[i], "-source="):
			source = strings.TrimPrefix(args[i], "-source=")
		default:
			filtered = append(filtered, args[i])
		}
	}

	if showVersion {
		fmt.Printf("digdir-cli %s\n", version)
		return
	}

	subcmd := ""
	if len(filtered) > 0 {
		subcmd = filtered[0]
	}

	switch subcmd {
	case "list":
		cmdList(source)
	case "install":
		cmdInstall(source, filtered[1:])
	case "update":
		cmdUpdate()
	case "version":
		fmt.Printf("digdir-cli %s\n", version)
	case "help":
		printUsage()
	default:
		cmdInteractive(source)
	}
}

func resolveSkills(sourcePath string) ([]skill.Skill, error) {
	src := &skill.Source{LocalPath: sourcePath}
	skillsDir, err := src.Resolve()
	if err != nil {
		return nil, fmt.Errorf("resolving skill source: %w", err)
	}
	skills, err := skill.Discover(skillsDir)
	if err != nil {
		return nil, fmt.Errorf("discovering skills: %w", err)
	}
	if len(skills) == 0 {
		return nil, fmt.Errorf("no skills found")
	}
	return skills, nil
}

func cmdInteractive(sourcePath string) {
	fmt.Println("Fetching skills...")
	skills, err := resolveSkills(sourcePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	_, err = tui.Run(skills)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func cmdList(sourcePath string) {
	skills, err := resolveSkills(sourcePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Available skills (%d):\n\n", len(skills))
	for _, s := range skills {
		fmt.Printf("  %-25s %s\n", s.Name, s.ShortDescription)
	}
}

func cmdInstall(sourcePath string, args []string) {
	fs := flag.NewFlagSet("install", flag.ExitOnError)
	fwFlag := fs.String("framework", "", "Agent framework (claude-code, cursor, copilot, codex)")
	skillsFlag := fs.String("skills", "", "Comma-separated list of skill names")
	targetFlag := fs.String("target", "", "Target repo path (use 'global' for global install, or a repo path)")
	fs.Parse(args)

	if *fwFlag == "" || *skillsFlag == "" || *targetFlag == "" {
		fmt.Fprintln(os.Stderr, "Usage: digdir-cli install --framework <fw> --skills <s1,s2> --target <path|global>")
		os.Exit(1)
	}

	fw := adapter.Framework(*fwFlag)
	if adapter.Get(fw) == nil {
		fmt.Fprintf(os.Stderr, "Unknown framework: %s\nSupported: claude-code, cursor, copilot, codex\n", *fwFlag)
		os.Exit(1)
	}

	allSkills, err := resolveSkills(sourcePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	requestedNames := strings.Split(*skillsFlag, ",")
	var selectedSkills []skill.Skill
	for _, name := range requestedNames {
		name = strings.TrimSpace(name)
		found := false
		for _, s := range allSkills {
			if s.Name == name {
				selectedSkills = append(selectedSkills, s)
				found = true
				break
			}
		}
		if !found {
			fmt.Fprintf(os.Stderr, "Skill not found: %s\n", name)
			os.Exit(1)
		}
	}

	cfg := installer.Config{
		Framework: fw,
		Skills:    selectedSkills,
	}

	if *targetFlag == "global" {
		cfg.Global = true
	} else {
		cfg.RepoPaths = []string{*targetFlag}
	}

	results := installer.Install(cfg)
	hasError := false
	for _, r := range results {
		if r.Err != nil {
			fmt.Fprintf(os.Stderr, "✗ %s: %v\n", r.Skill.Name, r.Err)
			hasError = true
		} else {
			fmt.Printf("✓ %s → %s\n", r.Skill.Name, r.TargetDir)
		}
	}
	if hasError {
		os.Exit(1)
	}
}

func cmdUpdate() {
	if err := updater.Update(version); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf(`digdir-cli %s - Provision agent skills from digdir/skills

Usage:
  digdir-cli                              Interactive TUI mode
  digdir-cli list                         List available skills
  digdir-cli install [flags]              Install skills non-interactively
  digdir-cli update                       Update to the latest release
  digdir-cli version                      Show version
  digdir-cli help                         Show this help

Global flags:
  --source <path>                         Use local skills repo instead of GitHub
  --version, -v                           Show version

Install flags:
  --framework <name>                      Agent framework (claude-code, cursor, copilot, codex)
  --skills <name,name,...>                Comma-separated skill names
  --target <path|global>                  Repo path or "global" for global install
`, version)
}
