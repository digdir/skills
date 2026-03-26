package tui

import (
	"fmt"
	"os"

	"github.com/digdir/skills/cli/internal/adapter"
	"github.com/digdir/skills/cli/internal/installer"
	"github.com/digdir/skills/cli/internal/skill"
)

// Run starts the interactive TUI and returns the installation results.
func Run(skills []skill.Skill) ([]installer.Result, error) {
	restore, err := rawMode()
	if err != nil {
		return nil, fmt.Errorf("enabling raw mode: %w", err)
	}
	defer restore()
	defer showCursor()
	hideCursor()

	// Step 1: Select framework
	fw, ok := selectFramework()
	if !ok {
		return nil, nil
	}

	// Step 2: Select skills
	selected, ok := selectSkills(skills)
	if !ok {
		return nil, nil
	}

	// Step 3: Select target
	global, repoPaths, ok := selectTarget(fw)
	if !ok {
		return nil, nil
	}

	// Step 4: Confirm
	ok = confirmInstall(fw, selected, global, repoPaths)
	if !ok {
		clearScreen()
		fmt.Println(dim("Cancelled."))
		return nil, nil
	}

	// Install
	cfg := installer.Config{
		Framework: fw,
		Skills:    selected,
		Global:    global,
		RepoPaths: repoPaths,
	}
	results := installer.Install(cfg)

	// Show results
	clearScreen()
	fmt.Println(bold("Results") + "\n")
	for _, r := range results {
		if r.Err != nil {
			fmt.Fprintf(os.Stderr, "%s %s: %v\n", red("✗"), r.Skill.Name, r.Err)
		} else {
			fmt.Printf("%s %s → %s\n", green("✓"), r.Skill.Name, dim(r.TargetDir))
		}
	}

	return results, nil
}

// header prints the app title.
func header() {
	fmt.Println(boldBlue("digdir-cli") + " " + dim("— Skill Provisioner") + "\n")
}

// selectFramework runs the framework selection step, returning the chosen framework.
// Returns false if the user cancelled.
func selectFramework() (adapter.Framework, bool) {
	frameworks := adapter.AllFrameworks()
	cursor := 0

	for {
		clearScreen()
		header()
		fmt.Println(bold("Select agent framework") + "\n")

		for i, fw := range frameworks {
			if i == cursor {
				fmt.Printf("  %s %s\n", yellow(">"), boldGreen(fw.DisplayName()))
			} else {
				fmt.Printf("    %s\n", fw.DisplayName())
			}
		}

		fmt.Println("\n" + dim("↑/↓ navigate • enter select • q quit"))

		key := readKey()
		switch key.event {
		case keyUp:
			if cursor > 0 {
				cursor--
			}
		case keyDown:
			if cursor < len(frameworks)-1 {
				cursor++
			}
		case keyEnter:
			return frameworks[cursor], true
		case keyCtrlC:
			return "", false
		case keyRune:
			if key.char == 'q' {
				return "", false
			}
		}
	}
}

// selectSkills runs the skill selection step.
func selectSkills(skills []skill.Skill) ([]skill.Skill, bool) {
	cursor := 0
	selected := make(map[int]bool)

	for {
		clearScreen()
		header()
		fmt.Println(bold("Select skills to install") + "\n")

		for i, sk := range skills {
			cur := "  "
			if i == cursor {
				cur = yellow("> ")
			}

			check := "[ ]"
			name := sk.Name
			if selected[i] {
				check = green("[x]")
				name = boldGreen(sk.Name)
			}

			fmt.Printf("  %s%s %s\n", cur, check, name)
			if sk.ShortDescription != "" {
				fmt.Printf("       %s\n", dim(sk.ShortDescription))
			}
		}

		count := 0
		for _, v := range selected {
			if v {
				count++
			}
		}

		fmt.Print("\n" + dim(fmt.Sprintf("↑/↓ navigate • space toggle • a all • enter confirm (%d selected) • esc back", count)) + "\n")

		key := readKey()
		switch key.event {
		case keyUp:
			if cursor > 0 {
				cursor--
			}
		case keyDown:
			if cursor < len(skills)-1 {
				cursor++
			}
		case keySpace:
			selected[cursor] = !selected[cursor]
		case keyEnter:
			if count > 0 {
				var result []skill.Skill
				for i, s := range skills {
					if selected[i] {
						result = append(result, s)
					}
				}
				return result, true
			}
		case keyEsc:
			return nil, false
		case keyCtrlC:
			return nil, false
		case keyRune:
			switch key.char {
			case 'a':
				allSelected := count == len(skills)
				for i := range skills {
					selected[i] = !allSelected
				}
			case 'q':
				return nil, false
			}
		}
	}
}

// selectTarget runs the install target selection step.
func selectTarget(fw adapter.Framework) (global bool, repoPaths []string, ok bool) {
	a := adapter.Get(fw)
	supportsGlobal := a != nil && a.GlobalPath() != ""
	cursor := 0

	for {
		clearScreen()
		header()
		fmt.Println(bold("Install target") + "\n")

		options := []string{}
		if supportsGlobal {
			globalDesc := ""
			if a != nil {
				globalDesc = " " + dim("("+a.GlobalPath()+")")
			}
			options = append(options, "Global"+globalDesc)
		}
		options = append(options, "Project repo(s)")

		for i, opt := range options {
			if i == cursor {
				fmt.Printf("  %s %s\n", yellow(">"), boldGreen(opt))
			} else {
				fmt.Printf("    %s\n", opt)
			}
		}

		fmt.Println("\n" + dim("↑/↓ navigate • enter select • esc back"))

		key := readKey()
		switch key.event {
		case keyUp:
			if cursor > 0 {
				cursor--
			}
		case keyDown:
			if cursor < len(options)-1 {
				cursor++
			}
		case keyEnter:
			if supportsGlobal && cursor == 0 {
				return true, nil, true
			}
			// Enter repo path input mode
			paths, inputOk := inputRepoPaths()
			if inputOk && len(paths) > 0 {
				return false, paths, true
			}
			// If cancelled or empty, stay on this screen
		case keyEsc:
			return false, nil, false
		case keyCtrlC:
			return false, nil, false
		}
	}
}

// inputRepoPaths lets the user type in repo paths one at a time.
func inputRepoPaths() ([]string, bool) {
	var paths []string
	var buf []byte

	for {
		clearScreen()
		header()
		fmt.Println(bold("Enter repo paths") + " " + dim("(one per line, empty line to finish)") + "\n")

		for i, p := range paths {
			fmt.Printf("  %s %s\n", green(fmt.Sprintf("%d.", i+1)), p)
		}

		prompt := fmt.Sprintf("  %d. ", len(paths)+1)
		fmt.Print(yellow(prompt) + cyan(string(buf)) + yellow("█"))

		fmt.Println("\n\n" + dim("enter add • enter (empty) finish • esc cancel"))

		key := readKey()
		switch key.event {
		case keyEnter:
			path := string(buf)
			if path == "" {
				if len(paths) > 0 {
					return paths, true
				}
				// No paths yet, ignore
			} else {
				paths = append(paths, path)
				buf = nil
			}
		case keyBackspace:
			if len(buf) > 0 {
				buf = buf[:len(buf)-1]
			} else if len(paths) > 0 {
				paths = paths[:len(paths)-1]
			}
		case keyEsc:
			return nil, false
		case keyCtrlC:
			return nil, false
		case keyRune:
			buf = append(buf, byte(key.char))
		case keySpace:
			buf = append(buf, ' ')
		}
	}
}

// confirmInstall shows a summary and asks for confirmation.
func confirmInstall(fw adapter.Framework, skills []skill.Skill, global bool, repoPaths []string) bool {
	cursor := 0 // 0 = install, 1 = cancel

	for {
		clearScreen()
		header()
		fmt.Println(bold("Confirm installation") + "\n")

		fmt.Printf("  Framework:  %s\n", boldGreen(fw.DisplayName()))

		fmt.Print("  Skills:     ")
		for i, s := range skills {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(green(s.Name))
		}
		fmt.Println()

		fmt.Print("  Target:     ")
		if global {
			a := adapter.Get(fw)
			if a != nil {
				fmt.Println(dim("global (" + a.GlobalPath() + ")"))
			}
		} else {
			for i, p := range repoPaths {
				if i > 0 {
					fmt.Print("              ")
				}
				fmt.Println(dim(p))
			}
		}

		fmt.Println()

		installBtn := "  [ Install ]  "
		cancelBtn := "  [ Cancel ]  "
		if cursor == 0 {
			installBtn = boldGreen("  [ Install ]  ")
		} else {
			cancelBtn = red("  [ Cancel ]  ")
		}
		fmt.Println(installBtn + cancelBtn)

		fmt.Println("\n" + dim("←/→ navigate • enter select • y/n shortcut • esc back"))

		key := readKey()
		switch key.event {
		case keyLeft:
			cursor = 0
		case keyRight:
			cursor = 1
		case keyEnter:
			return cursor == 0
		case keyEsc:
			return false
		case keyCtrlC:
			return false
		case keyRune:
			switch key.char {
			case 'y':
				return true
			case 'n':
				return false
			}
		}
	}
}
